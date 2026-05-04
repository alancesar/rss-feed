package ml

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

var (
	ErrEmbeddingNotFound = errors.New("embedding not found")
)

type (
	Embedding struct {
		Model  string
		Vector []float32
	}

	Properties struct {
		Model string
		URL   string
	}

	embeddingRequest struct {
		Model string `json:"model"`
		Input string `json:"input"`
	}

	embeddingResponse struct {
		Model      string      `json:"model"`
		Embeddings [][]float32 `json:"embeddings"`
	}

	OllamaClient struct {
		client *http.Client
		props  Properties
	}
)

func NewOllamaClient(client *http.Client, props Properties) *OllamaClient {
	return &OllamaClient{
		client: client,
		props:  props,
	}
}

func (c *OllamaClient) GetEmbeddings(ctx context.Context, text string) (Embedding, error) {
	body, err := json.Marshal(embeddingRequest{
		Model: c.props.Model,
		Input: text,
	})
	if err != nil {
		return Embedding{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.props.URL, bytes.NewBuffer(body))
	if err != nil {
		return Embedding{}, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return Embedding{}, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	var output embeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&output); err != nil {
		return Embedding{}, err
	}

	if len(output.Embeddings) == 0 {
		return Embedding{}, ErrEmbeddingNotFound
	}

	return Embedding{
		Model:  output.Model,
		Vector: output.Embeddings[0],
	}, nil
}
