// Package ml provides clients for machine learning services.
package ml

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

var (
	// ErrEmbeddingNotFound is returned when the service answers successfully but with no embedding.
	ErrEmbeddingNotFound = errors.New("embedding not found")
)

type (
	// Embedding is the vector representation of a text, along with the model that produced it.
	Embedding struct {
		Model  string
		Vector []float32
	}

	// Properties configures an OllamaClient.
	Properties struct {
		// Model is the name of the embedding model to request.
		Model string
		// URL is the full address of the embedding endpoint.
		URL string
	}

	// embeddingRequest is the JSON body sent to the embedding endpoint.
	embeddingRequest struct {
		Model string `json:"model"`
		Input string `json:"input"`
	}

	// embeddingResponse is the JSON body returned by the embedding endpoint.
	embeddingResponse struct {
		Model      string      `json:"model"`
		Embeddings [][]float32 `json:"embeddings"`
	}

	// OllamaClient is an HTTP client for the Ollama embedding API.
	OllamaClient struct {
		client *http.Client
		props  Properties
	}
)

// NewOllamaClient creates an OllamaClient that sends its requests through client, using
// the model and endpoint from props.
func NewOllamaClient(client *http.Client, props Properties) *OllamaClient {
	return &OllamaClient{
		client: client,
		props:  props,
	}
}

// GetEmbeddings POSTs the text as JSON to the configured endpoint and returns the first
// embedding of the response. Non-2xx statuses are returned as errors that include the response
// body, and a response without embeddings yields ErrEmbeddingNotFound.
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

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return Embedding{}, fmt.Errorf("ollama returned non-2xx status %d (could not read body: %w)", resp.StatusCode, readErr)
		}
		return Embedding{}, fmt.Errorf("ollama returned non-2xx status %d: %s", resp.StatusCode, string(body))
	}

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
