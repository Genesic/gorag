package text

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Genesic/gorag"
	"github.com/Genesic/gorag/loader"
)

type Loader struct {
	opts loader.Options
}

func NewLoader(opts ...loader.Options) loader.Loader {
	o := loader.DefaultLoaderOptions()
	if len(opts) > 0 {
		o = opts[0]
	}
	return &Loader{opts: o}
}

// Load 從檔案路徑載入純文字文檔
func (l *Loader) Load(ctx context.Context, source string) (*gorag.Document, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// 檢查檔案資訊
	info, err := os.Stat(source)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// 檢查檔案大小限制
	if l.opts.MaxSize > 0 && info.Size() > l.opts.MaxSize {
		return nil, fmt.Errorf("file size %d exceeds max size %d", info.Size(), l.opts.MaxSize)
	}

	// 讀取檔案內容
	content, err := os.ReadFile(source)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// 產生文檔 ID（使用檔案路徑的 hash）
	absPath, err := filepath.Abs(source)
	if err != nil {
		absPath = source
	}
	id := generateID(absPath)

	// 建立 Document
	doc := gorag.Document{
		ID:      id,
		Content: string(content),
		Metadata: gorag.Metadata{
			Source:    absPath,
			Title:     filepath.Base(source),
			CreatedAt: info.ModTime(),
			Extra:     make(map[string]any),
		},
	}

	// 附加額外的 metadata
	for k, v := range l.opts.Metadata {
		doc.Metadata.Extra[k] = v
	}

	return &doc, nil
}

// LoadMultiple 批次載入多個檔案
func (l *Loader) LoadMultiple(ctx context.Context, sources []string) ([]gorag.Document, error) {
	var docs []gorag.Document

	for _, source := range sources {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		loaded, err := l.Load(ctx, source)
		if err != nil {
			return nil, fmt.Errorf("failed to load %s: %w", source, err)
		}
		docs = append(docs, *loaded)
	}

	return docs, nil
}

// LoadURL 從 URL 載入純文字內容
func (l *Loader) LoadURL(ctx context.Context, url string) (*gorag.Document, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// 檢查內容大小限制
	if l.opts.MaxSize > 0 && resp.ContentLength > l.opts.MaxSize {
		return nil, fmt.Errorf("content size %d exceeds max size %d", resp.ContentLength, l.opts.MaxSize)
	}

	var content []byte
	if l.opts.MaxSize > 0 {
		content, err = io.ReadAll(io.LimitReader(resp.Body, l.opts.MaxSize))
	} else {
		content, err = io.ReadAll(resp.Body)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	id := generateID(url)

	doc := gorag.Document{
		ID:      id,
		Content: string(content),
		Metadata: gorag.Metadata{
			Source:    url,
			Title:     url,
			CreatedAt: time.Now(),
			Extra:     make(map[string]any),
		},
	}

	for k, v := range l.opts.Metadata {
		doc.Metadata.Extra[k] = v
	}

	return &doc, nil
}

func generateID(source string) string {
	hash := sha256.Sum256([]byte(source + time.Now().String()))
	return hex.EncodeToString(hash[:16])
}

// 確保實作了 loader.Loader interface
var _ loader.Loader = (*Loader)(nil)
