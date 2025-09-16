package models

import (
	"encoding/json"
	"time"

	"github.com/openai/openai-go/packages/param"
	"github.com/openai/openai-go/responses"
)

// Primary template Ai template object, used for cloning the template from
// public templates and creating new private templates
type TenantAiTemplate struct {
	ID                *string          `json:"id"`
	SourceID          *string          `json:"source_id"` // New field
	UserID            *string          `json:"user_id"`
	TenantID          *string          `json:"tenant_id"`
	Title             *string          `json:"title"`
	Description       *string          `json:"description"` // Renamed from 'details'
	Category          *string          `json:"category"`
	AiVendor          *string          `json:"ai_vendor"` // New field
	AiModel           *string          `json:"ai_model"`  // New field
	Configuration     *AiConfiguration `json:"configuration"`
	Privacy           *bool            `json:"privacy"`            // New field
	OriginalPublisher *string          `json:"original_publisher"` // Original publisher field will only be populated when the template is copied from a public template
	PublishedBy       *string          `json:"published_by"`       // Published by is only populated when the template is copied from a public template
	CreatedAt         time.Time        `json:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at"`
	FirstName         *string          `json:"first_name"`
	LastName          *string          `json:"last_name"`
	UserPicture       *string          `json:"user_picture"`
}

type SpAiTemplate struct {
	ID            *string          `json:"id"`
	SourceID      *string          `json:"source_id"` // New field
	UserID        *string          `json:"user_id"`
	Title         *string          `json:"title"`
	Description   *string          `json:"description"` // Renamed from 'details'
	Category      *string          `json:"category"`
	AiVendor      *string          `json:"ai_vendor"` // New field
	AiModel       *string          `json:"ai_model"`  // New field
	Configuration *AiConfiguration `json:"configuration"`
	PublishedBy   *string          `json:"published_by"` // Published by is only populated when the template is copied from a public template
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
}

// TODO: Need to update it (both BE and FE) as some of the fields are not available in Responses API
type AiConfiguration struct {
	AiTemperature float64 `json:"ai_temperature"` // Temperature value
	PromptRole    string  `json:"prompt_role"`    // Always system
	SystemConfig  string  `json:"system_config"`  // This is the system configuration, prompt engineering
	TopP          float64 `json:"top_p"`
	MaxTokens     int     `json:"max_tokens"`
}

type DocumentSelection struct {
	ID   string `json:"id"`
	Type string `json:"type" default:"doc-fragment"`
	Data string `json:"data"`
}

type PromptResource struct {
	ID             string              `json:"id"`
	UserID         string              `json:"user_id"`
	TenantID       string              `json:"tenant_id"`
	ConversationID string              `json:"conversation_id"`
	PromptRole     string              `json:"prompt_role"`
	Prompt         string              `json:"prompt"`
	Selections     []DocumentSelection `json:"selections"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
}

// This is the object send into the Prompt endpoint
type PromptRequest struct {
	UserID         string              `json:"user_id"`
	TenantID       string              `json:"tenant_id"`
	ConversationID string              `json:"conversation_id"`
	Message        string              `json:"message"`
	Selections     []DocumentSelection `json:"selections"`
	CreatedAt      *time.Time          `json:"created_at"`
}

type CompletionRequest struct {
	UserID           string      `json:"user_id"`
	TenantID         string      `json:"tenant_id"`
	Prompt           string      `json:"prompt"`
	Tools            interface{} `json:"tools"`
	ConversationID   string      `json:"conversation_id"`
	PromptTokens     int32       `json:"prompt_tokens"`
	CompletionTokens int32       `json:"completion_tokens"`
	TotalTokens      int32       `json:"total_tokens"`
	CreatedAt        *time.Time  `json:"created_at"`
}

type CompletionRequestResource struct {
	ID               string    `json:"id"`
	UserID           string    `json:"user_id"`
	TenantID         string    `json:"tenant_id"`
	ConversationID   string    `json:"conversation_id"`
	PromptRole       string    `json:"prompt_role"`
	Prompt           string    `json:"prompt"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	PromptTokens     int       `json:"prompt_tokens"`
	CompletionTokens int       `json:"completion_tokens"`
	TotalTokens      int       `json:"total_tokens"`
}

type TenantTokenUsageRequest struct {
	UserID           string      `json:"user_id"`
	TenantID         string      `json:"tenant_id"`
	ConversationID   *string     `json:"conversation_id"`
	AiVendor         string      `json:"ai_vendor"`
	AiModel          string      `json:"ai_model"`
	Tools            interface{} `json:"tools"`
	PromptTokens     int32       `json:"prompt_tokens"`
	CompletionTokens int32       `json:"completion_tokens"`
}

type TenantTokenUsageResource struct {
	TenantTokenUsageRequest
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

type MessageResource struct {
	ID         string              `json:"id"`
	Role       string              `json:"role"`
	CreatedAt  time.Time           `json:"created_at"`
	Content    string              `json:"content"`
	Selections []DocumentSelection `json:"selections"`
}

// type Message struct {
// 	Role    string `json:"role"`
// 	Content string `json:"content"`
// }

type OpenAiConfig struct {
	OpenAIProjectID string          `json:"openai_project_id"`
	OpenAIApiKey    string          `json:"openai_api_key"`
	OpenAIApiKeyID  string          `json:"openai_api_key_id"`
	Assistants      json.RawMessage `json:"assistants"`
	VectorStores    json.RawMessage `json:"vector_stores"`
}

// AssistantParams is a custom implementation of responses.ResponseNewParams that properly handles JSON unmarshaling
// While responses.ResponseNewParams contains the same fields, its unmarshaling logic fails to properly decode
// the JSON response from the API. This struct provides identical functionality with correct JSON handling.
type AssistantParams struct {
	// Model ID used to generate the response, like `gpt-4o` or `o1`. OpenAI offers a
	// wide range of models with different capabilities, performance characteristics,
	// and price points. Refer to the
	// [model guide](https://platform.openai.com/docs/models) to browse and compare
	// available models.
	Model responses.ResponsesModel `json:"model,omitzero"`
	// Inserts a system (or developer) message as the first item in the model's
	// context.
	//
	// When using along with `previous_response_id`, the instructions from a previous
	// response will be not be carried over to the next response. This makes it simple
	// to swap out system (or developer) messages in new responses.
	Instructions param.Opt[string] `json:"instructions,omitzero"`
	// An upper bound for the number of tokens that can be generated for a response,
	// including visible output tokens and
	// [reasoning tokens](https://platform.openai.com/docs/guides/reasoning).
	MaxOutputTokens param.Opt[int64] `json:"max_output_tokens,omitzero"`
	// Whether to allow the model to run tool calls in parallel.
	ParallelToolCalls param.Opt[bool] `json:"parallel_tool_calls,omitzero"`
	// What sampling temperature to use, between 0 and 2. Higher values like 0.8 will
	// make the output more random, while lower values like 0.2 will make it more
	// focused and deterministic. We generally recommend altering this or `top_p` but
	// not both.
	Temperature param.Opt[float64] `json:"temperature,omitzero"`
	// An alternative to sampling with temperature, called nucleus sampling, where the
	// model considers the results of the tokens with top_p probability mass. So 0.1
	// means only the tokens comprising the top 10% probability mass are considered.
	//
	// We generally recommend altering this or `temperature` but not both.
	TopP param.Opt[float64] `json:"top_p,omitzero"`
	// Specify additional output data to include in the model response. Currently
	// supported values are:
	//
	//   - `file_search_call.results`: Include the search results of the file search tool
	//     call.
	//   - `message.input_image.image_url`: Include image urls from the input message.
	//   - `computer_call_output.output.image_url`: Include image urls from the computer
	//     call output.
	Include []responses.ResponseIncludable `json:"include,omitzero"`
	// Set of 16 key-value pairs that can be attached to an object. This can be useful
	// for storing additional information about the object in a structured format, and
	// querying for objects via API or the dashboard.
	//
	// Keys are strings with a maximum length of 64 characters. Values are strings with
	// a maximum length of 512 characters.
	Metadata responses.MetadataParam `json:"metadata,omitzero"`
	// The truncation strategy to use for the model response.
	//
	//   - `auto`: If the context of this response and previous ones exceeds the model's
	//     context window size, the model will truncate the response to fit the context
	//     window by dropping input items in the middle of the conversation.
	//   - `disabled` (default): If a model response will exceed the context window size
	//     for a model, the request will fail with a 400 error.
	//
	// Any of "auto", "disabled".
	Truncation responses.ResponseNewParamsTruncation `json:"truncation,omitzero"`
	// **o-series models only**
	//
	// Configuration options for
	// [reasoning models](https://platform.openai.com/docs/guides/reasoning).
	Reasoning responses.ReasoningParam `json:"reasoning,omitzero"`

	// Forces assistant to return data in JSON object format
	OutputJSONObject param.Opt[bool] `json:"output_json_object,omitzero"`

	FileSearch    *responses.FileSearchToolParam `json:"file_search,omitzero"`
	WebSearch     *responses.WebSearchToolParam  `json:"web_search,omitzero"`
	FunctionCalls []responses.FunctionToolParam  `json:"function_calls,omitzero"`
}

type ChatContext struct {
	UserID         string
	TenantID       string
	ConversationID *string

	ResourceIdentifier
}
