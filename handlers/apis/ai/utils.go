package ai

import (
	"encoding/json"
	"fmt"
	"log"
	"sententiawebapi/handlers/apis/tenantManagement"
	"sententiawebapi/handlers/models"
	"sententiawebapi/utilities"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/responses"
)

func GetOpenAiClient(tenantID string) (*openai.Client, *models.OpenAiConfig, error) {
	openAiConfig, err := utilities.GetOpenAiConfig(tenantManagement.DB, tenantID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get OpenAI config: %w", err)
	}

	client := openai.NewClient(
		option.WithAPIKey(openAiConfig.OpenAIApiKey),
	)

	return &client, openAiConfig, nil
}

func mergeAssistantParams(responseParams *responses.ResponseNewParams, assistantParams *models.AssistantParams, openAiConfig *models.OpenAiConfig) {
	if assistantParams.Model != "" {
		responseParams.Model = assistantParams.Model
	}

	if assistantParams.Instructions.IsPresent() {
		responseParams.Instructions = assistantParams.Instructions
	}

	if len(assistantParams.Include) > 0 {
		responseParams.Include = assistantParams.Include
	}

	responseParams.Metadata = assistantParams.Metadata

	if assistantParams.Reasoning.IsPresent() {
		responseParams.Reasoning = assistantParams.Reasoning
	}

	if assistantParams.MaxOutputTokens.IsPresent() {
		responseParams.MaxOutputTokens = assistantParams.MaxOutputTokens
	}

	if assistantParams.ParallelToolCalls.IsPresent() {
		responseParams.ParallelToolCalls = assistantParams.ParallelToolCalls
	}

	if assistantParams.Temperature.IsPresent() {
		responseParams.Temperature = assistantParams.Temperature
	}

	if assistantParams.TopP.IsPresent() {
		responseParams.TopP = assistantParams.TopP
	}

	if assistantParams.Truncation != "" {
		responseParams.Truncation = assistantParams.Truncation
	}

	if assistantParams.FileSearch != nil && openAiConfig != nil {
		var vectorStores map[string]string
		err := json.Unmarshal(openAiConfig.VectorStores, &vectorStores)

		if err != nil {
			log.Printf("Failed to unmarshal vector stores: %v", err)
		} else {
			for i, vectorStoreName := range assistantParams.FileSearch.VectorStoreIDs {
				assistantParams.FileSearch.VectorStoreIDs[i] = vectorStores[vectorStoreName]
			}

			responseParams.Tools = append(responseParams.Tools, responses.ToolUnionParam{
				OfFileSearch: assistantParams.FileSearch,
			})
		}

	}

	if assistantParams.WebSearch != nil {
		responseParams.Tools = append(responseParams.Tools, responses.ToolUnionParam{
			OfWebSearch: assistantParams.WebSearch,
		})
	}

	if assistantParams.FunctionCalls != nil {
		for _, functionCallConfig := range assistantParams.FunctionCalls {
			responseParams.Tools = append(responseParams.Tools, responses.ToolUnionParam{
				OfFunction: &functionCallConfig,
			})
		}
	}

	if assistantParams.OutputJSONObject.Value {
		responseParams.Text = responses.ResponseTextConfigParam{
			Format: responses.ResponseFormatTextConfigUnionParam{
				OfJSONObject: &responses.ResponseFormatJSONObjectParam{},
			},
		}
	}
}
