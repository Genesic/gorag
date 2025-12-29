# GoRAG Framework Interface 設計文檔

## 概述

本文檔描述 GoRAG framework 所需的所有 interface 設計，採用依賴注入模式，讓使用者可以自行實現並注入各個組件。

---

## Interface 總覽

| Interface | 套件 | 功能描述 |
|-----------|------|----------|
| Loader | `loader` | 從各種來源載入文檔 |
| Splitter | `splitter` | 將文檔分割成小塊 (chunks) |
| Embedder | `embedder` | 將文本轉換成向量 |
| VectorStore | `vectorstore` | 向量的存儲與相似度搜尋 |
| Retriever | `retriever` | 根據查詢檢索相關文檔 |
| Reranker | `reranker` | 重新排序檢索結果 |
| LLM | `llm` | 語言模型生成回答 |
| PromptTemplate | `prompt` | 提示詞模板管理 |

---

## 詳細設計

### 1. Loader

**套件**: `loader`

**功能**: 從各種來源載入原始文檔，轉換成統一的 Document 格式。

**典型實現**:
- TextLoader: 載入純文字檔案
- PDFLoader: 載入 PDF 文件
- HTMLLoader: 載入網頁內容
- JSONLoader: 載入 JSON 資料
- DirectoryLoader: 批次載入整個目錄

**方法**:
| 方法 | 說明 |
|------|------|
| `Load(ctx, source string) ([]Document, error)` | 從指定來源載入文檔 |
| `LoadMultiple(ctx, sources []string) ([]Document, error)` | 批次載入多個來源 |

---

### 2. Splitter

**套件**: `splitter`

**功能**: 將大型文檔分割成適合嵌入和檢索的小塊 (chunks)。

**典型實現**:
- CharacterSplitter: 按字元數分割
- RecursiveCharacterSplitter: 遞迴分割（優先保持段落完整）
- TokenSplitter: 按 token 數分割
- SentenceSplitter: 按句子分割
- MarkdownSplitter: 按 Markdown 結構分割

**方法**:
| 方法 | 說明 |
|------|------|
| `Split(ctx, docs []Document) ([]Document, error)` | 將文檔分割成 chunks |

**設定參數**:
- `ChunkSize`: 每個 chunk 的大小
- `ChunkOverlap`: chunk 之間的重疊大小

---

### 3. Embedder

**套件**: `embedder`

**功能**: 將文本轉換成向量表示 (embedding)，用於後續的相似度計算。

**典型實現**:
- OpenAIEmbedder: 使用 OpenAI embedding API
- OllamaEmbedder: 使用本地 Ollama 模型

**方法**:
| 方法 | 說明 |
|------|------|
| `Embed(ctx, texts []string) ([][]float32, error)` | 批次嵌入多個文本 |
| `EmbedQuery(ctx, query string) ([]float32, error)` | 嵌入單一查詢（某些模型對查詢有特殊處理）|
| `Dimension() int` | 返回向量維度 |

---

### 4. VectorStore

**套件**: `vectorstore`

**功能**: 存儲文檔向量，並提供相似度搜尋功能。

**典型實現**:
- MemoryStore: 記憶體內存儲（適合小型資料集）
- PgVectorStore: PostgreSQL + pgvector
- QdrantStore: Qdrant 向量資料庫

**方法**:
| 方法 | 說明 |
|------|------|
| `Add(ctx, docs []Document, embeddings [][]float32) error` | 新增文檔和向量 |
| `Delete(ctx, ids []string) error` | 刪除指定文檔 |
| `Search(ctx, embedding []float32, opts SearchOptions) ([]SearchResult, error)` | 向量相似度搜尋 |
| `Clear(ctx) error` | 清空所有資料 |

**SearchOptions**:
- `TopK`: 返回前 K 個最相似的結果
- `ScoreThreshold`: 最低相似度閾值
- `Filter`: 元數據過濾條件

---

### 5. Retriever

**套件**: `retriever`

**功能**: 高層次的檢索介面，整合 Embedder 和 VectorStore，提供端到端的檢索功能。

**典型實現**:
- VectorRetriever: 基於向量相似度的檢索
- KeywordRetriever: 基於關鍵字的檢索（BM25）
- HybridRetriever: 混合檢索（向量 + 關鍵字）
- MultiQueryRetriever: 生成多個查詢變體後合併結果
- ContextualCompressionRetriever: 壓縮檢索結果，只保留相關部分

**方法**:
| 方法 | 說明 |
|------|------|
| `Retrieve(ctx, query string, opts RetrieveOptions) ([]Document, error)` | 檢索相關文檔 |

**RetrieveOptions**:
- `TopK`: 返回數量
- `Filter`: 元數據過濾條件

---

### 6. Reranker

**套件**: `reranker`

**功能**: 對初步檢索結果進行重新排序，提高相關性。

**典型實現**:
- CohereReranker: 使用 Cohere rerank API
- CrossEncoderReranker: 使用 cross-encoder 模型
- LLMReranker: 使用 LLM 評估相關性

**方法**:
| 方法 | 說明 |
|------|------|
| `Rerank(ctx, query string, docs []Document, topN int) ([]Document, error)` | 重新排序並返回前 N 個 |

---

### 7. LLM

**套件**: `llm`

**功能**: 語言模型介面，用於生成最終回答。

**典型實現**:
- OpenAILLM: OpenAI GPT 系列
- AnthropicLLM: Anthropic Claude 系列
- OllamaLLM: 本地 Ollama 模型
- AzureOpenAILLM: Azure OpenAI 服務

**方法**:
| 方法 | 說明 |
|------|------|
| `Generate(ctx, prompt string, opts GenerateOptions) (string, error)` | 生成文本 |
| `GenerateStream(ctx, prompt string, opts GenerateOptions) (<-chan StreamChunk, error)` | 串流生成 |
| `Chat(ctx, messages []Message, opts GenerateOptions) (Message, error)` | 對話模式 |
| `ChatStream(ctx, messages []Message, opts GenerateOptions) (<-chan Message, error)` | 對話串流模式 |

**GenerateOptions**:
- `MaxTokens`: 最大 token 數
- `Temperature`: 生成溫度
- `TopP`: nucleus sampling
- `StopWords`: 停止詞

---

### 8. PromptTemplate

**套件**: `prompt`

**功能**: 管理和格式化提示詞模板。

**典型實現**:
- GoTemplate: 使用 Go text/template


**方法**:
| 方法 | 說明 |
|------|------|
| `Format(ctx, vars map[string]any) (string, error)` | 用變數填充模板 |
| `InputVariables() []string` | 返回模板需要的變數名稱 |

---

## 核心資料結構

### Document

```go
type Document struct {
    ID        string            // 唯一識別碼
    Content   string            // 文本內容
    Metadata  Metadata          // 元數據
    Embedding []float32         // 向量（可選）
}
```

### Metadata

```go
type Metadata struct {
    Source      string         // 來源
    Title       string         // 標題
    ChunkIndex  int            // chunk 索引
    TotalChunks int            // 總 chunk 數
    ParentID    string         // 父文檔 ID
    Extra       map[string]any // 自定義欄位
}
```

### SearchResult

```go
type SearchResult struct {
    Document Document
    Score    float32  // 相似度分數
}
```

### Message (用於對話)

```go
type Message struct {
    Role    string  // "system", "user", "assistant"
    Content string
}
```

---

## 建議目錄結構

```
gorag/
├── document.go           # Document 結構定義
├── errors.go             # 統一錯誤定義
│
├── loader/
│   ├── loader.go         # Loader interface
│   ├── text.go           # TextLoader 實現
│   └── pdf.go            # PDFLoader 實現
│
├── splitter/
│   ├── splitter.go       # Splitter interface
│   └── recursive.go      # RecursiveCharacterSplitter 實現
│
├── embedder/
│   ├── embedder.go       # Embedder interface
│   └── openai.go         # OpenAI 實現
│
├── vectorstore/
│   ├── store.go          # VectorStore interface
│   ├── memory.go         # MemoryStore 實現
│   └── pgvector.go       # PgVector 實現
│
├── retriever/
│   ├── retriever.go      # Retriever interface
│   └── vector.go         # VectorRetriever 實現
│
├── reranker/
│   ├── reranker.go       # Reranker interface
│   └── cohere.go         # Cohere 實現
│
├── llm/
│   ├── llm.go            # LLM interface
│   ├── openai.go         # OpenAI 實現
│   └── anthropic.go      # Anthropic 實現
│
├── prompt/
│   ├── template.go       # PromptTemplate interface
│   └── string.go         # StringTemplate 實現
│
└── rag/
    └── rag.go            # RAG 主流程整合
```

---

## RAG 流程圖

```
┌─────────┐    ┌──────────┐    ┌──────────┐    ┌─────────────┐
│ Loader  │───▶│ Splitter │───▶│ Embedder │───▶│ VectorStore │
└─────────┘    └──────────┘    └──────────┘    └─────────────┘
                                                      │
                                                      ▼
┌─────┐    ┌──────────────┐    ┌──────────┐    ┌───────────┐
│ LLM │◀───│PromptTemplate│◀───│ Reranker │◀───│ Retriever │
└─────┘    └──────────────┘    └──────────┘    └───────────┘
    │
    ▼
┌──────────┐
│ Response │
└──────────┘
```

**Indexing 階段**: Loader → Splitter → Embedder → VectorStore

**Query 階段**: Retriever → Reranker → PromptTemplate → LLM → Response