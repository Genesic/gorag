package recursive

import (
	"context"
	"strings"

	"github.com/Genesic/gorag"
	"github.com/Genesic/gorag/splitter"
)

// Splitter 遞迴字元分割器
// 會依序嘗試不同的分隔符來分割文本，優先保持段落和句子的完整性
type Splitter struct {
	opts splitter.Options
}

// New 建立新的 RecursiveCharacterSplitter
func NewSplitter(opts ...splitter.Options) splitter.Splitter {
	o := splitter.DefaultOptions()
	if len(opts) > 0 {
		o = opts[0]
	}
	return &Splitter{opts: o}
}

// Split 將文檔分割成 chunks
func (s *Splitter) Split(ctx context.Context, docs []gorag.Document) ([]gorag.Document, error) {
	var result []gorag.Document

	for _, doc := range docs {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		chunks := s.splitText(doc.Content, s.opts.Separators)
		totalChunks := len(chunks)

		for i, chunk := range chunks {
			newDoc := gorag.Document{
				ID:      doc.ID + "_chunk_" + itoa(i),
				Content: chunk,
				Metadata: gorag.Metadata{
					Source:      doc.Metadata.Source,
					Title:       doc.Metadata.Title,
					Author:      doc.Metadata.Author,
					CreatedAt:   doc.Metadata.CreatedAt,
					ChunkIndex:  i,
					TotalChunks: totalChunks,
					ParentID:    doc.ID,
					Extra:       copyExtra(doc.Metadata.Extra),
				},
			}
			result = append(result, newDoc)
		}
	}

	return result, nil
}

// splitText 遞迴分割文本
func (s *Splitter) splitText(text string, separators []string) []string {
	if len(text) == 0 {
		return nil
	}

	// 如果文本已經小於 chunk size，直接返回
	if len(text) <= s.opts.ChunkSize {
		return []string{strings.TrimSpace(text)}
	}

	// 如果沒有分隔符了，按字元強制分割
	if len(separators) == 0 {
		return s.splitBySize(text)
	}

	separator := separators[0]
	remainingSeparators := separators[1:]

	// 嘗試用當前分隔符分割
	var splits []string
	if separator == "" {
		// 空分隔符表示按字元分割
		for _, r := range text {
			splits = append(splits, string(r))
		}
	} else {
		splits = strings.Split(text, separator)
	}

	// 合併分割後的片段
	var chunks []string
	var currentChunk strings.Builder

	for i, split := range splits {
		split = strings.TrimSpace(split)
		if split == "" {
			continue
		}

		// 如果單個片段就超過 chunk size，需要遞迴用下一個分隔符分割
		if len(split) > s.opts.ChunkSize {
			// 先保存當前累積的 chunk
			if currentChunk.Len() > 0 {
				chunks = append(chunks, strings.TrimSpace(currentChunk.String()))
				currentChunk.Reset()
			}
			// 遞迴分割這個大片段
			subChunks := s.splitText(split, remainingSeparators)
			chunks = append(chunks, subChunks...)
			continue
		}

		// 計算加入這個片段後的大小
		newSize := currentChunk.Len() + len(split)
		if currentChunk.Len() > 0 && separator != "" {
			newSize += len(separator)
		}

		if newSize > s.opts.ChunkSize {
			// 當前 chunk 已滿，保存並開始新的 chunk
			if currentChunk.Len() > 0 {
				chunks = append(chunks, strings.TrimSpace(currentChunk.String()))
			}
			// 處理 overlap：從上一個 chunk 取一部分
			currentChunk.Reset()
			if s.opts.ChunkOverlap > 0 && len(chunks) > 0 {
				lastChunk := chunks[len(chunks)-1]
				overlapStart := len(lastChunk) - s.opts.ChunkOverlap
				if overlapStart < 0 {
					overlapStart = 0
				}
				currentChunk.WriteString(lastChunk[overlapStart:])
				if separator != "" {
					currentChunk.WriteString(separator)
				}
			}
		}

		// 加入當前片段
		if currentChunk.Len() > 0 && separator != "" && i > 0 {
			currentChunk.WriteString(separator)
		}
		currentChunk.WriteString(split)
	}

	// 保存最後一個 chunk
	if currentChunk.Len() > 0 {
		chunks = append(chunks, strings.TrimSpace(currentChunk.String()))
	}

	return chunks
}

// splitBySize 按固定大小強制分割
func (s *Splitter) splitBySize(text string) []string {
	var chunks []string
	runes := []rune(text)

	for i := 0; i < len(runes); i += s.opts.ChunkSize - s.opts.ChunkOverlap {
		end := i + s.opts.ChunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunk := string(runes[i:end])
		if trimmed := strings.TrimSpace(chunk); trimmed != "" {
			chunks = append(chunks, trimmed)
		}
		if end == len(runes) {
			break
		}
	}

	return chunks
}

// itoa 簡單的整數轉字串
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var digits []byte
	for i > 0 {
		digits = append([]byte{byte('0' + i%10)}, digits...)
		i /= 10
	}
	return string(digits)
}

// copyExtra 複製 extra metadata
func copyExtra(extra map[string]any) map[string]any {
	if extra == nil {
		return make(map[string]any)
	}
	copied := make(map[string]any, len(extra))
	for k, v := range extra {
		copied[k] = v
	}
	return copied
}
