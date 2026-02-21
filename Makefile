# --- 變數設定 ---
DOCKER_USER ?= julianshen
IMAGE_NAME = alyasan
PORT = 9000

# 模型版本控制
MODEL_SIZE ?= 4b
ifeq ($(MODEL_SIZE), 12b)
    MODEL_TAG = translategemma:12b
else
    MODEL_TAG = translategemma:4b
endif

# 完整的 Docker 標籤 (例如: yourusername/gemma-translator:4b)
FULL_IMAGE_TAG = $(DOCKER_USER)/$(IMAGE_NAME):$(MODEL_SIZE)

.PHONY: build run stop push login help

help:
	@echo "可用指令:"
	@echo "  make build MODEL_SIZE=4b   - 構建鏡像 (預設 4b)"
	@echo "  make push  MODEL_SIZE=4b   - 推送至 Docker Hub"
	@echo "  make run   MODEL_SIZE=4b   - 本地啟動測試"
	@echo "  make login                 - 登入 Docker Hub"

# 1. 登入 Docker Hub
login:
	docker login

# 2. 構建鏡像 (使用 Debian-Slim 優化版)
build:
	@echo "正在構建 $(FULL_IMAGE_TAG)..."
	docker build --build-arg MODEL_TAG=$(MODEL_TAG) -t $(FULL_IMAGE_TAG) .

# 3. 推送鏡像
# 注意：模型鏡像很大 (5GB+)，請確保網路環境穩定
push:
	@echo "正在推送 $(FULL_IMAGE_TAG) 到 Docker Hub..."
	docker push $(FULL_IMAGE_TAG)

# 4. 運行本地測試
run:
	@echo "啟動服務於 http://localhost:$(PORT)"
	docker run -d \
		--name alyasan \
		-p $(PORT):3000 \
		--env OLLAMA_KEEP_ALIVE=-1 \
		$(FULL_IMAGE_TAG)

stop:
	docker stop alyasan || true
	docker rm alyasan || true