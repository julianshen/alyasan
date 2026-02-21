package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ollama/ollama/api"
)

// 將靜態檔案編譯進執行檔
//
//go:embed static/*
var staticFiles embed.FS

var (
	detectedModel = ""
	ollamaClient  *api.Client
	translateTpl  *template.Template
)

type TranslateRequest struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Text   string `json:"text"`
}

func init() {
	var err error
	ollamaClient, err = api.ClientFromEnvironment()
	if err != nil {
		log.Fatal("❌ 無法初始化 Ollama 客戶端:", err)
	}
	translateTpl = template.Must(template.New("translate").Parse(`You are a professional {{.Source}} to {{.Target}} translator. Your goal is to accurately convey the meaning and nuances of the original {{.Source}} text while adhering to {{.Target}} grammar, vocabulary, and cultural sensitivities.
Produce only the {{.Target}} translation, without any additional explanations or commentary. Please translate the following {{.Source}} text into {{.Target}}:

{{.Text}}`))
}

// 自動偵測本地已下載的 TranslateGemma 模型
func detectModel() {
	ctx := context.Background()
	log.Println("🔍 正在偵測本地 TranslateGemma 模型...")

	for i := 0; i < 10; i++ { // 重試 10 次
		list, err := ollamaClient.List(ctx)
		if err == nil {
			for _, m := range list.Models {
				if strings.Contains(m.Name, "translategemma") {
					detectedModel = m.Name
					log.Printf("✅ 成功對接模型: %s", detectedModel)
					return
				}
			}
		}
		time.Sleep(2 * time.Second)
	}
	log.Println("⚠️ 未偵測到模型，將在請求時嘗試使用預設 translategemma:4b")
	detectedModel = "translategemma:4b"
}

func main() {
	// 非同步偵測模型，不阻塞啟動
	go detectModel()

	// 1. 靜態網頁路由
	content, _ := fs.Sub(staticFiles, "static")
	http.Handle("/", http.FileServer(http.FS(content)))

	// 2. 翻譯 API (SSE 流式)
	http.HandleFunc("/api/translate", handleTranslate)

	// 3. 系統資訊 API (供前端顯示模型版本)
	http.HandleFunc("/api/info", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"model": detectedModel})
	})

	port := ":3000"
	log.Printf("🌐 翻譯服務已啟動: http://localhost%s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}

func handleTranslate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "只支援 POST", http.StatusMethodNotAllowed)
		return
	}

	// 1. 確保模型已經掛載成功
	if detectedModel == "" {
		http.Error(w, "AI 模型尚未就緒，請稍後再試", http.StatusServiceUnavailable)
		return
	}

	var req TranslateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "無效的 JSON", http.StatusBadRequest)
		return
	}

	var prompt strings.Builder
	if err := translateTpl.Execute(&prompt, req); err != nil {
		http.Error(w, "生成提示詞失敗", http.StatusInternalServerError)
		return
	}

	// 3. 設定 SSE Header (防快取優化)
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	stream := true
	genReq := &api.GenerateRequest{
		Model:  detectedModel,
		Prompt: prompt.String(),
		Stream: &stream,
		Options: map[string]interface{}{
			"temperature": 0.0, // 官方建議：翻譯需要 100% 確定性
			"num_ctx":     4096,
		},
	}

	// 4. 執行串流
	err := ollamaClient.Generate(r.Context(), genReq, func(resp api.GenerateResponse) error {
		if resp.Response == "" {
			return nil
		}
		fmt.Fprintf(w, "data: %s\n\n", resp.Response)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		return nil
	})

	if err != nil {
		// 優雅處理客戶端中斷
		if r.Context().Err() == context.Canceled {
			log.Printf("ℹ️ 使用者取消了翻譯請求")
		} else {
			log.Printf("❌ 翻譯出錯: %v", err)
		}
	}
}
