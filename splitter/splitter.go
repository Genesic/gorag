package splitter

import (
	"context"

	"github.com/Genesic/gorag"
)

// Splitter 定義文檔分割器的介面
// 用於將大型文檔分割成適合嵌入和檢索的小塊 (chunks)
type Splitter interface {
	// Split 將文檔分割成 chunks
	Split(ctx context.Context, docs []gorag.Document) ([]gorag.Document, error)
}

// Options 分割器通用選項
type Options struct {
	// ChunkSize 每個 chunk 的目標大小（字元數）
	ChunkSize int

	// ChunkOverlap chunk 之間的重疊大小（字元數）
	ChunkOverlap int

	// Separators 分隔符列表，按優先順序排列
	// RecursiveCharacterSplitter 會依序嘗試這些分隔符
	Separators []string
}

// DefaultOptions 返回預設的分割器選項
func DefaultOptions() Options {
	return Options{
		ChunkSize:    1000,
		ChunkOverlap: 200,
		Separators: []string{
			"\n\n", // 段落
			"\n",   // 換行
			"。",   // 中文句號
			".",    // 英文句號
			"！",   // 中文驚嘆號
			"!",    // 英文驚嘆號
			"？",   // 中文問號
			"?",    // 英文問號
			"；",   // 中文分號
			";",    // 英文分號
			"，",   // 中文逗號
			",",    // 英文逗號
			" ",    // 空格
			"",     // 單字元（最後手段）
		},
	}
}
