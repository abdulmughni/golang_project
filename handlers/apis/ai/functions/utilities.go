package aiFunctions

import (
	"context"
	"fmt"
	"strings"

	openai "github.com/openai/openai-go"
)

func embedQuery(openaiClient *openai.Client, query string) (*string, error) {
	embedResp, err := openaiClient.Embeddings.New(context.Background(), openai.EmbeddingNewParams{
		Model: openai.EmbeddingModelTextEmbedding3Small,
		Input: openai.EmbeddingNewParamsInputUnion{
			OfString: openai.String(query),
		},
	})
	if err != nil || len(embedResp.Data) == 0 {
		return nil, fmt.Errorf("failed to generate embedding for a query message: %v", err)
	}

	embedding := embedResp.Data[0].Embedding
	vectorStr := "'[" + strings.Trim(strings.ReplaceAll(fmt.Sprint(embedding), " ", ","), "[]") + "]'"

	return &vectorStr, nil
}
