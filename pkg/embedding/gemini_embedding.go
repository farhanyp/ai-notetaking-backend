package embedding

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type EmbeddingRequestContentPart struct {
	Text string `json:"text"`
}

type EmbeddingRequestContent struct {
	Parts []EmbeddingRequestContentPart `json:"parts"`
}

type EmbeddingRequest struct {
	Model    string                  `json:"model"`
	Content  EmbeddingRequestContent `json:"content"`
	TaskType string                  `json:"task_type"`
}

type EmbeddingResponseEmbedding struct {
	Values []float32 `json:"values"`
}

type EmbeddingResponse struct {
	Embedding EmbeddingResponseEmbedding `json:"embedding"`
}

func GetGeminiEmbedding(
	apiKey string,
	text string,
	task_type string,
) (*EmbeddingResponse, error) {

	geminiReq := EmbeddingRequest{
		Model: "models/gemini-embedding-exp-03-07",
		Content: EmbeddingRequestContent{
			Parts: []EmbeddingRequestContentPart{
				{
					Text: text,
				},
			},
		},
		TaskType: task_type,
	}

	geminiReqJson, err := json.Marshal(geminiReq)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"Post",
		"https://generativelanguage.googleapis.com/v1beta/models/gemini-embedding-001:embedContent",
		bytes.NewBuffer(geminiReqJson),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("x-goog-api-key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	resBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error from response, code %d, body %s", res.StatusCode, resBytes)
	}

	var resEmbedding EmbeddingResponse
	err = json.Unmarshal(resBytes, &resEmbedding)
	if err != nil {
		return nil, err
	}

	return &resEmbedding, nil

}