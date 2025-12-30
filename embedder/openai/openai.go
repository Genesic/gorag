package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Genesic/gorag/embedder"
)

const (
	defaultBaseURL = "https://api.openai.com/v1"

	ModelTextEmbedding3Small = "text-embedding-3-small"
	ModelTextEmbedding3Large = "text-embedding-3-large"
	ModelTextEmbeddingAda002 = "text-embedding-ada-002"

	defaultModel = ModelTextEmbedding3Small
)

// Model 維度對照
var modelDimensions = map[string]int{
	ModelTextEmbedding3Small: 1536,
	ModelTextEmbedding3Large: 3072,
	ModelTextEmbeddingAda002: 1536,
}

// Embedder OpenAI 嵌入器
type Embedder struct {
	apiKey     string
	baseURL    string
	model      string
	dimensions int
	batchSize  int
	client     *http.Client
}

// Options OpenAI 嵌入器選項
type Options struct {
	// APIKey OpenAI API 金鑰（必填）
	APIKey string

	// BaseURL API 基礎 URL（可選，預設為 OpenAI 官方 API）
	BaseURL string

	// Model 模型名稱（可選，預設為 text-embedding-3-small）
	Model string

	// Dimensions 向量維度（可選，部分模型支援自訂維度）
	Dimensions int

	// BatchSize 批次處理大小（可選，預設為 100）
	BatchSize int

	// HTTPClient 自訂 HTTP 客戶端（可選）
	HTTPClient *http.Client
}

func NewEmbedder(opts Options) (embedder.Embedder, error) {
	if opts.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	baseURL := opts.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	model := opts.Model
	if model == "" {
		model = defaultModel
	}

	dimensions := opts.Dimensions
	if dimensions == 0 {
		if dim, ok := modelDimensions[model]; ok {
			dimensions = dim
		} else {
			dimensions = 1536 // 預設維度
		}
	}

	batchSize := opts.BatchSize
	if batchSize <= 0 {
		batchSize = 100
	}

	client := opts.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	return &Embedder{
		apiKey:     opts.APIKey,
		baseURL:    baseURL,
		model:      model,
		dimensions: dimensions,
		batchSize:  batchSize,
		client:     client,
	}, nil
}

// Embed 批次嵌入多個文本
func (e *Embedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	var allEmbeddings [][]float32

	// 分批處理
	for i := 0; i < len(texts); i += e.batchSize {
		end := i + e.batchSize
		if end > len(texts) {
			end = len(texts)
		}
		batch := texts[i:end]

		embeddings, err := e.embedBatch(ctx, batch)
		if err != nil {
			return nil, fmt.Errorf("failed to embed batch %d-%d: %w", i, end, err)
		}
		allEmbeddings = append(allEmbeddings, embeddings...)
	}

	return allEmbeddings, nil
}

// EmbedQuery 嵌入單一查詢
func (e *Embedder) EmbedQuery(ctx context.Context, query string) ([]float32, error) {
	embeddings, err := e.embedBatch(ctx, []string{query})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}
	return embeddings[0], nil
}

// Dimension 返回向量維度
func (e *Embedder) Dimension() int {
	return e.dimensions
}

// embedBatch 內部方法：批次呼叫 API
func (e *Embedder) embedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	reqBody := embeddingRequest{
		Model: e.model,
		Input: texts,
	}

	// 如果模型支援自訂維度
	if e.model == ModelTextEmbedding3Small || e.model == ModelTextEmbedding3Large {
		reqBody.Dimensions = e.dimensions
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.baseURL+"/embeddings", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.apiKey)

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var embResp embeddingResponse
	if err = json.Unmarshal(body, &embResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// 按 index 排序並提取 embeddings
	embeddings := make([][]float32, len(texts))
	for _, data := range embResp.Data {
		if data.Index < len(embeddings) {
			embeddings[data.Index] = data.Embedding
		}
	}

	return embeddings, nil
}

// API 請求/回應結構

type embeddingRequest struct {
	Model      string   `json:"model"`
	Input      []string `json:"input"`
	Dimensions int      `json:"dimensions,omitempty"`
}

type embeddingResponse struct {
	Data  []embeddingData `json:"data"`
	Model string          `json:"model"`
	Usage usageInfo       `json:"usage"`
}

type embeddingData struct {
	Index     int       `json:"index"`
	Embedding []float32 `json:"embedding"`
}

type usageInfo struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}
