# Alyasan

Alyasan is a lightweight local translation web app powered by Ollama and TranslateGemma. It serves a simple HTML UI and streams translations over SSE for fast, incremental results.

## Features
- Local, offline-capable translation via Ollama
- Streaming responses using Server-Sent Events (SSE)
- Simple bilingual UI with preset languages
- Auto-detects installed `translategemma` model variants

## Requirements
- Go 1.25+ (for local development)
- Ollama installed and running
- A TranslateGemma model, e.g. `translategemma:4b` or `translategemma:12b`

## Run Locally
1. Start Ollama and ensure the model is available:
   - `ollama pull translategemma:4b`
2. Start the server:
   - `go run .`
3. Open the UI:
   - `http://localhost:3000`

## Docker
Build and run using the provided `Makefile` (defaults to `MODEL_SIZE=4b`):

- Build image:
  - `make build MODEL_SIZE=4b`
- Run container (maps host port 9000 to container 3000):
  - `make run MODEL_SIZE=4b`
- Stop container:
  - `make stop`

## Configuration
- Model detection happens at startup; if none is found, it falls back to `translategemma:4b`.
- Server listens on `:3000`.

## API
- `POST /api/translate`
  - Body:
    ```json
    {"source":"English (eng_Latn)","target":"Traditional Chinese (zho_Hant)","text":"Hello"}
    ```
  - Response: SSE stream of `data: ...` chunks

- `GET /api/info`
  - Returns current model name

## Supported Languages (UI)
English, Traditional Chinese, Simplified Chinese, Japanese, Korean, Arabic, Indonesian, Hindi, Polish, German, French, Spanish, Italian, Russian, Vietnamese, Thai.

## Notes
- The frontend is embedded from `static/index.html` using `go:embed`.
- Arabic output switches to RTL direction in the UI.
