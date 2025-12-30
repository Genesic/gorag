package embedder

import (
	"context"
)

// Embedder 定義文本嵌入器的介面
// 用於將文本轉換成向量表示 (embedding)
type Embedder interface {
	// Embed 批次嵌入多個文本
	Embed(ctx context.Context, texts []string) ([][]float32, error)

	// EmbedQuery 嵌入單一查詢
	// 某些模型對查詢有特殊處理（如加入特殊 prefix）
	EmbedQuery(ctx context.Context, query string) ([]float32, error)

	// Dimension 返回向量維度
	Dimension() int
}

// Options 嵌入器通用選項
type Options struct {
	// Model 模型名稱
	Model string

	// BatchSize 批次處理大小
	BatchSize int

	// Dimensions 向量維度（部分模型支援自訂維度）
	Dimensions int
}

// DefaultOptions 返回預設選項
func DefaultOptions() Options {
	return Options{
		BatchSize: 100,
	}
}