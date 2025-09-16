package aiFunctions

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sententiawebapi/handlers/models"
	"strings"

	openai "github.com/openai/openai-go"
	"github.com/openai/openai-go/responses"
	"golang.org/x/sync/errgroup"
)

func GetFunctionDefinitions(chatCtx *models.ChatContext) []responses.ToolUnionParam {
	functionTools := make([]responses.ToolUnionParam, 0)

	functionTools = append(functionTools, responses.ToolUnionParam{
		OfFunction: &responses.FunctionToolParam{
			Name:        "document_search",
			Description: openai.String("Semantic search for documents within the project. Returns only (small) fragments of document(s). Each fragment starts with index and document title."),
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]string{
						"type":        "string",
						"description": "Natural-language search phrase. If you want current/concrete document, leave empty and use scope.",
					},
					"limit": map[string]string{
						"type": "number",
						"description": `
							Maximum number of matching fragments to return.
							Each fragment corresponds to a content block (e.g., paragraph, table, heading, or list), similar to Notion-style editors.
						`,
					},
					"scope": map[string]any{
						"type": "array",
						"items": map[string]string{
							"type": "string",
						},
						"description": `
							Optional list of document IDs to restrict the search scope.
							Use ["current"] to search only the current document.
							If user input includes references (e.g. @ref(doc: <uuid>)), include those IDs only if they are relevant to the query.
							Leave empty or omit to search the entire project.
						`,
					},
				},
				"required": []string{"query", "limit"},
			},
		},
	})

	functionTools = append(functionTools, responses.ToolUnionParam{
		OfFunction: &responses.FunctionToolParam{
			Name: "diagram_search",
			Description: openai.String(`
				Semantic search for diagrams within the project.
			`),
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]string{
						"type":        "string",
						"description": "Natural-language search phrase. If you want current diagram, leave empty and use scope.",
					},
					"limit": map[string]string{
						"type": "number",
						"description": `
							Max diagrams to return
						`,
					},
					"scope": map[string]any{
						"type": "array",
						"items": map[string]string{
							"type": "string",
						},
						"description": `
							Optional list of diagram UUIDs to restrict the search scope.
							Use ["current"] to retrieve currently active diagram.
							If user input includes references (e.g. @ref(diag: <uuid>)), include those IDs only if they are relevant to the query.
							Leave empty or omit to search the entire project.
						`,
					},
				},
				"required": []string{"query", "limit"},
			},
		},
	})

	if chatCtx.ResourceGroupID != nil {
		functionTools = append(functionTools, responses.ToolUnionParam{
			OfFunction: &responses.FunctionToolParam{
				Name: "project_info",
				Description: openai.String(`
					Retrieves high-level information about the current project, including its title, status, category, complexity, description, and all associated requirements.
					This information can help you better understand the project context before generating ideas, giving suggestions, or answering questions.
				`),
			},
		})
	}

	return functionTools
}

func ExecuteFunctionCall(openai *openai.Client, chatCtx *models.ChatContext, functionName string, rawArguments string) (string, error) {
	if functionName == "document_search" {
		var arguments DocSearchRequest
		if err := json.Unmarshal([]byte(rawArguments), &arguments); err != nil {
			return "", fmt.Errorf("invalid function arguments: %v", rawArguments)
		}

		results, err := DocumentSearch(openai, &arguments, chatCtx)
		if err != nil {
			return "", fmt.Errorf("Document search failed: %v", err)
		}

		if len(results) == 0 {
			return "", nil
		}

		var sb strings.Builder
		for i, result := range results {
			sb.WriteString(fmt.Sprintf("#%d Document: %s\n", i+1, result.Title))
			sb.WriteString(result.Content)
			sb.WriteString("\n\n")
		}

		sb.WriteString("Note: If some expected content is missing, it may not have been indexed yet.\n")
		sb.WriteString("Feel free to ask the user for clarification or to paste the relevant content here.")
		sb.WriteString("\n\n")

		log.Print(sb.String())

		return sb.String(), nil
	}

	if functionName == "diagram_search" {
		var arguments DiagramSearchRequest
		if err := json.Unmarshal([]byte(rawArguments), &arguments); err != nil {
			return "", fmt.Errorf("invalid function arguments: %v", rawArguments)
		}

		results, err := DiagramSearch(openai, &arguments, chatCtx)
		if err != nil {
			return "", fmt.Errorf("Diagram search failed: %v", err)
		}

		if len(results) == 0 {
			return "", nil
		}

		var sb strings.Builder

		sb.WriteString("How to read diagram:\n")
		sb.WriteString("[<group>] #<x-position order L->R> <node label> (icon filename) // node definition\n")
		sb.WriteString("--> #<target order> <target label> // connection to another node\n")
		sb.WriteString("--<connection icon>--> #<order> <label>\n\n")

		for i, result := range results {
			sb.WriteString(fmt.Sprintf("#%d ", i+1))
			sb.WriteString(result.Content)
			sb.WriteString("\n\n")
		}

		sb.WriteString("Note: If some expected content is missing, it may not have been indexed yet.\n")
		sb.WriteString("\n\n")

		log.Print(sb.String())

		return sb.String(), nil
	}

	if functionName == "project_info" {
		projectInfo, err := GetProjectInfo(chatCtx)
		if err != nil {
			return "", fmt.Errorf("Project Info function call failed: %v", err)
		}

		log.Printf("Project info: %s", projectInfo)

		return projectInfo, nil
	}

	return "", fmt.Errorf("unknown function call: %s", functionName)
}

func ExecuteFunctionCallsParallel(
	ctx context.Context,
	client *openai.Client,
	chatCtx *models.ChatContext,
	outputs []responses.ResponseOutputItemUnion,
) ([]responses.ResponseInputItemUnionParam, error) {
	type task struct {
		index int
		out   responses.ResponseOutputItemUnion
	}
	var tasks []task
	for i, o := range outputs {
		if o.Type == "function_call" {
			tasks = append(tasks, task{i, o})
		}
	}
	if len(tasks) == 0 {
		return nil, nil // nothing to do
	}

	results := make([]responses.ResponseInputItemUnionParam, len(tasks))

	// errgroup with context (cancels all task when one fails)
	g, _ := errgroup.WithContext(ctx)

	sem := make(chan struct{}, 8)

	for _, t := range tasks {
		t := t // capture
		g.Go(func() error {
			sem <- struct{}{}        // acquire
			defer func() { <-sem }() // release

			out, err := ExecuteFunctionCall(
				client, chatCtx, t.out.Name, t.out.Arguments,
			)
			if err != nil {
				return fmt.Errorf("%s failed: %w", t.out.Name, err)
			}

			results[t.index] = responses.
				ResponseInputItemParamOfFunctionCallOutput(t.out.CallID, out)
			return nil
		})
	}

	// waiting for all or first error
	if err := g.Wait(); err != nil {
		return nil, err
	}

	return results, nil
}
