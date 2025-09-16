package ai

import (
	"context"
	"log"
	"net/http"
	"sententiawebapi/handlers/apis/tenantManagement"
	"sententiawebapi/handlers/models"
	"sententiawebapi/utilities"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"
	"github.com/openai/openai-go/responses"
)

type TextRequestPayload struct {
	CollapseToEnd bool    `json:"collapseToEnd"`
	Format        string  `json:"format"`
	Html          bool    `json:"html"`
	Stream        bool    `json:"stream"`
	Text          string  `json:"text"`
	Tone          *string `json:"tone"`
}

type OutputType string

const (
	OutputTypeHTML  OutputType = "html"
	OutputTypePlain OutputType = "plain"
)

var systemInstructionsOutline = []string{
	"General instructions:",
	"You are a helpful assistant used for tasks within a text editor.",
	"The document is a technical documentation created by an IT architect.",
	"Always treat input as a fragment of HTML or plain text extracted from the document.",
	"Non-html characters (especially markdown formatting) are prohibited.",
	"Output must be ready to insert into a document.",
	"Be lazy: prefer producing less text over more text, unless the task explicitly asks for longer content.",
	"\n\n",
}

func getRequestPayload(c *gin.Context) (*TextRequestPayload, error) {
	var reqPayload TextRequestPayload
	if err := c.ShouldBindJSON(&reqPayload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return nil, err
	}
	return &reqPayload, nil
}

func streamChunkedResponse(c *gin.Context, text string, aiConfiguration *models.AssistantParams, outputType OutputType) {
	userID, tenantID, ok := utilities.ProcessIdentity(c)
	if !ok {
		return
	}

	openAiConfig, err := utilities.GetOpenAiConfig(tenantManagement.DB, tenantID)
	if err != nil {
		log.Printf("Failed to get OpenAI config: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	client := openai.NewClient(
		option.WithAPIKey(openAiConfig.OpenAIApiKey),
	)

	var responseParams = &responses.ResponseNewParams{
		Model:           "gpt-4o-mini",
		User:            openai.String(userID),
		Temperature:     openai.Float(0.5),
		TopP:            openai.Float(1),
		MaxOutputTokens: openai.Int(4000),
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(text),
		},
	}
	mergeAssistantParams(responseParams, aiConfiguration, nil)

	stream := client.Responses.NewStreaming(context.Background(), *responseParams)

	// Set headers for chunked streaming
	c.Header("Content-Type", "text/plain; charset=UTF-8") // Tiptap expects plain text streaming, even if it contains HTML.
	c.Header("Transfer-Encoding", "chunked")

	// Write the HTTP status code before streaming begins
	c.Writer.WriteHeader(http.StatusOK)

	// Make sure the ResponseWriter supports streaming by checking for the Flusher interface.
	writer := c.Writer
	flusher, ok := writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Streaming not supported"})
		return
	}

	var started bool
	var ended bool

	// Stream out the response tokens one by one
	for stream.Next() {
		event := stream.Current()
		token := event.Delta

		if event.JSON.Text.IsPresent() {
			ended = true
		}

		if outputType == OutputTypeHTML && !started && !ended {
			writer.Write([]byte("<body>\n"))
			started = true
		}

		if !ended {
			writer.Write([]byte(token))
			flusher.Flush()
		}
	}

	if err := stream.Err(); err != nil {
		log.Printf("Error during streaming: %v", err)
	}

	// Finalize the HTML output
	if outputType == OutputTypeHTML {
		writer.Write([]byte("\n</body>"))
		flusher.Flush()
	}

	// Save token usage after the stream is finished
	defer func() {
		response := stream.Current().Response

		err := newTokenUsageResource(&models.TenantTokenUsageRequest{
			TenantID:         tenantID,
			UserID:           userID,
			ConversationID:   nil,
			AiVendor:         "openai",
			AiModel:          response.Model,
			Tools:            map[string]interface{}{},
			PromptTokens:     int32(response.Usage.InputTokens),
			CompletionTokens: int32(response.Usage.OutputTokens),
		})
		if err == nil {
			log.Printf("Token usage stored successfully")
			log.Printf("Total usage tokens: %v", response.Usage.TotalTokens)
		}
	}()
}

func constructInstructions(args ...interface{}) param.Opt[string] {
	var allLines []string

	for _, arg := range args {
		switch v := arg.(type) {
		case string:
			allLines = append(allLines, v)
		case []string:
			allLines = append(allLines, v...)
		}
	}

	return openai.String(strings.Join(allLines, " "))
}

func SimplifyTextHandler(c *gin.Context) {
	reqPayload, err := getRequestPayload(c)
	if err != nil {
		return
	}

	streamChunkedResponse(c, reqPayload.Text, &models.AssistantParams{
		Instructions: constructInstructions(
			systemInstructionsOutline,
			"Task: Simplify given document fragment.",
			"I will manually wrap your response in <body></body> tags, don't include them yourself.",
		),
	}, OutputTypeHTML)
}

func FixSpellingAndGrammarHandler(c *gin.Context) {
	reqPayload, err := getRequestPayload(c)
	if err != nil {
		return
	}

	streamChunkedResponse(c, reqPayload.Text, &models.AssistantParams{
		Instructions: constructInstructions(
			systemInstructionsOutline,
			"Task: fix spelling and grammar in the given fragment.",
			"Focus on fixing mistakes, keeping the html format and text meaning intact.",
			"I will manually wrap your response in <body></body> tags, don't include them yourself.",
		),
	}, OutputTypeHTML)
}

func ShortenTextHandler(c *gin.Context) {
	reqPayload, err := getRequestPayload(c)
	if err != nil {
		return
	}

	streamChunkedResponse(c, reqPayload.Text, &models.AssistantParams{
		Instructions: constructInstructions(
			systemInstructionsOutline,
			"Task: shorten given fragment, preserving the original meaning.",
			"I will manually wrap your response in <body></body> tags, don't include them yourself.",
		),
	}, OutputTypeHTML)
}

func ExtendTextHandler(c *gin.Context) {
	reqPayload, err := getRequestPayload(c)
	if err != nil {
		return
	}

	streamChunkedResponse(c, reqPayload.Text, &models.AssistantParams{
		Instructions: constructInstructions(
			systemInstructionsOutline,
			"Task: make the given fragment longer, preserving the original meaning.",
			"I will manually wrap your response in <body></body> tags, don't include them yourself.",
		),
	}, OutputTypeHTML)
}

func AdjustToneHandler(c *gin.Context) {
	reqPayload, err := getRequestPayload(c)
	if err != nil {
		return
	}

	if reqPayload.Tone == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tone is required"})
		return
	}

	streamChunkedResponse(c, reqPayload.Text, &models.AssistantParams{
		Instructions: constructInstructions(
			systemInstructionsOutline,
			"Task: adjust the tone of the given fragment to `"+*reqPayload.Tone+"`.",
			"I will manually wrap your response in <body></body> tags, don't include them yourself.",
		),
	}, OutputTypeHTML)
}

func TldrHandler(c *gin.Context) {
	reqPayload, err := getRequestPayload(c)
	if err != nil {
		return
	}

	streamChunkedResponse(c, reqPayload.Text, &models.AssistantParams{
		Instructions: constructInstructions(
			systemInstructionsOutline,
			"Task: provide a TL;DR (simple, concise summary) of the given fragment.",
			"I will manually wrap your response in <body></body> tags, don't include them yourself.",
		),
	}, OutputTypeHTML)
}

func AiWriterHandler(c *gin.Context) {
	reqPayload, err := getRequestPayload(c)
	if err != nil {
		return
	}

	instructions := []string{
		"Task: generate a detailed HTML response based on the provided user instructions.",
		"The input is plain text, and you should expand it into a comprehensive paragraph or more.",
		"Focus on writing everything relevant about the specified subject, unless instructed by user otherwise.",
		"Respond with well-structured HTML content.",
		"I will manually wrap your response in <body></body> tags, don't include them yourself.",
		"When tasked with code generation, use highlight-js syntax and html elements; Always wrap code blocks with <pre> and <code> tags;",
		"Specify correct language like this: <code class='language-...'",
		"NEVER EVER use markdown formatting, especially ```html or ```powershell",
	}

	if reqPayload.Tone != nil {
		instructions = append(instructions, "Write in the tone of: `"+*reqPayload.Tone+"`.")
	}

	streamChunkedResponse(c, reqPayload.Text, &models.AssistantParams{
		Instructions: constructInstructions(systemInstructionsOutline, instructions),
	}, OutputTypeHTML)
}

func AutocompleteTextHandler(c *gin.Context) {
	var reqPayload struct {
		Text string `json:"text"`
	}
	if err := c.ShouldBindJSON(&reqPayload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	streamChunkedResponse(c, reqPayload.Text, &models.AssistantParams{
		Instructions: constructInstructions(
			systemInstructionsOutline,
			"Task: autocomplete given fragment.",
			"The input is plain text, consisting of one or a few sentences before the user's current cursor position.",
			"Provide a relatively short completion for the sentence or a natural follow-up.",
			"Provide only a few words, it's important to keep the text short.",
			"Respond with plain text only, without any formatting.",
		),
	}, OutputTypePlain)
}

// This is a special case for generating descriptions
var systemInstructionsOutlineProjectDescription = []string{
	"General instructions:",
	"You are a helpful assistant used to generate project descriptions from very little context.",
	"You will be provided a project title and a short description of the project and you will try to generate a detailed description of the project.",
	"Always stay concise and to the point, make sure that you generate a description that is as relevant as possible to the project title and short description.",
	"Non-html characters (especially markdown formatting) are prohibited.",
	"Output must be ready to insert into a textarea field.",
	"Response should be under 150 words.",
	"\n\n",
}

func ProjectDescriptionHandler(c *gin.Context) {
	reqPayload, err := getRequestPayload(c)
	if err != nil {
		return
	}

	streamChunkedResponse(c, reqPayload.Text, &models.AssistantParams{
		Instructions: constructInstructions(
			systemInstructionsOutlineProjectDescription,
			"Task: generate a detailed description of the project.",
			"I will manually wrap your response in <body></body> tags, don't include them yourself.",
		),
	}, OutputTypeHTML)
}
