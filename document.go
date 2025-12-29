package gorag

import "time"

// Document 代表 RAG 系統中的文檔單元
// 可以是原始文檔，也可以是分割後的 chunk
type Document struct {
	// ID 文檔的唯一識別碼
	ID string

	// Content 文檔的文本內容
	Content string

	// Metadata 文檔的元數據
	Metadata Metadata

	// Embedding 文檔的向量表示（可選，由 Embedder 填充）
	Embedding []float32
}

// Metadata 文檔元數據
type Metadata struct {
	// Source 文檔來源（檔案路徑、URL 等）
	Source string

	// Title 文檔標題
	Title string

	// Author 作者
	Author string

	// CreatedAt 創建時間
	CreatedAt time.Time

	// ChunkIndex 如果是分割後的 chunk，表示其索引位置
	ChunkIndex int

	// TotalChunks 原文檔分割後的總 chunk 數
	TotalChunks int

	// ParentID 如果是 chunk，指向原文檔的 ID
	ParentID string

	// Extra 其他自定義元數據
	Extra map[string]any
}

// SearchResult 搜尋結果
type SearchResult struct {
	Document Document
	Score    float32
}

// NewDocument 創建新的 Document
func NewDocument(id, content string) *Document {
	return &Document{
		ID:      id,
		Content: content,
		Metadata: Metadata{
			Extra: make(map[string]any),
		},
	}
}

// WithSource 設定文檔來源
func (d *Document) WithSource(source string) *Document {
	d.Metadata.Source = source
	return d
}

// WithTitle 設定文檔標題
func (d *Document) WithTitle(title string) *Document {
	d.Metadata.Title = title
	return d
}

// WithMetadata 設定完整元數據
func (d *Document) WithMetadata(meta Metadata) *Document {
	d.Metadata = meta
	if d.Metadata.Extra == nil {
		d.Metadata.Extra = make(map[string]any)
	}
	return d
}

// SetExtra 設定額外的元數據
func (d *Document) SetExtra(key string, value any) *Document {
	if d.Metadata.Extra == nil {
		d.Metadata.Extra = make(map[string]any)
	}
	d.Metadata.Extra[key] = value
	return d
}

// GetExtra 取得額外的元數據
func (d *Document) GetExtra(key string) (any, bool) {
	if d.Metadata.Extra == nil {
		return nil, false
	}
	v, ok := d.Metadata.Extra[key]
	return v, ok
}

// Clone 深拷貝 Document
func (d *Document) Clone() *Document {
	clone := &Document{
		ID:      d.ID,
		Content: d.Content,
		Metadata: Metadata{
			Source:      d.Metadata.Source,
			Title:       d.Metadata.Title,
			Author:      d.Metadata.Author,
			CreatedAt:   d.Metadata.CreatedAt,
			ChunkIndex:  d.Metadata.ChunkIndex,
			TotalChunks: d.Metadata.TotalChunks,
			ParentID:    d.Metadata.ParentID,
			Extra:       make(map[string]any),
		},
		Embedding: make([]float32, len(d.Embedding)),
	}
	copy(clone.Embedding, d.Embedding)

	for k, v := range d.Metadata.Extra {
		clone.Metadata.Extra[k] = v
	}
	return clone
}
