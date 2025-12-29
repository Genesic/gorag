package loader

import (
	"context"

	"github.com/Genesic/gorag"
)

// Loader 定義文檔載入器的介面
// 用於從各種來源（檔案、URL、資料庫等）載入文檔
type Loader interface {
	Load(ctx context.Context, source string) (*gorag.Document, error)
	LoadURL(ctx context.Context, url string) (*gorag.Document, error)
	LoadMultiple(ctx context.Context, sources []string) ([]gorag.Document, error)
}

type Options struct {
	// Encoding 文件編碼，預設為 UTF-8
	Encoding string

	// MaxSize 最大檔案大小（bytes），0 表示不限制
	MaxSize int64

	// Metadata 額外的元數據，會附加到所有載入的文檔
	Metadata map[string]any
}

// DefaultLoaderOptions 返回預設的載入器選項
func DefaultLoaderOptions() Options {
	return Options{
		Encoding: "UTF-8",
		MaxSize:  0,
		Metadata: make(map[string]any),
	}
}
