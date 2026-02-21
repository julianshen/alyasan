# --- 第一階段：編譯 Go 程式 (保持不變) ---
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /alyasan main.go

# --- 第二階段：輕量化運行環境 ---
FROM debian:bookworm-slim

# 安裝必要依賴：ca-certificates 用於 HTTPS 下載模型
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl ca-certificates sudo procps zstd && \
    rm -rf /var/lib/apt/lists/*

# 【修正處】直接下載 Ollama 二進制檔並移動到路徑中
# 這樣可以避開 install.sh 腳本在容器內執行時的權限與系統服務偵測問題
RUN curl -fsSL https://ollama.com/install.sh | OLLAMA_INSTALL_ONLY=1 sh

WORKDIR /app

# 複製 Go 執行檔
COPY --from=builder /alyasan .

# 接收模型版本參數
ARG MODEL_TAG=translategemma:4b

# 預下載模型
# 注意：Ollama 需要一個 HOME 目錄來存放模型，Debian 預設已有 /root
RUN OLLAMA_HOST=127.0.0.1:11434 ollama serve & \
    echo "Waiting for Ollama..." && \
    until curl -s http://127.0.0.1:11434/api/tags > /dev/null; do sleep 2; done && \
    echo "Server is up, pulling ${MODEL_TAG}..." && \
    ollama pull ${MODEL_TAG} && \
    pkill ollama && \
    sleep 2

EXPOSE 3000

ENV OLLAMA_HOST=0.0.0.0
ENV OLLAMA_KEEP_ALIVE=-1

# 啟動腳本
ENTRYPOINT ["sh", "-c", "ollama serve & sleep 2 && ./alyasan"]