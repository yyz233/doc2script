# doc2script backend

Converts novel chapters into TV drama scripts using LLM.

## Quick Start

```bash
# Required: configure LLM backend
export LLM_BASE_URL="http://your-llm-server/v1"
export LLM_MODEL_NAME_LC="your-model-name"
export LLM_API_KEY="your-api-key"

# Optional
export PORT=8081
export MAX_CONCURRENCY=5

go run main.go
```

## API

### Health check

```
GET /health
```

### Upload file

```
POST /api/v1/upload
Content-Type: multipart/form-data

fields:
  file: document to upload (.docx, .txt, .doc)
```

Response: `200` on success
```json
{
  "success": true,
  "message": "file uploaded successfully",
  "files": ["example.docx"],
  "saved_path": "/path/to/saved/file"
}
```

### Generate scripts (async)

```
POST /api/v1/generate
Content-Type: application/json

{
  "model": "lc",
  "gentype": "d",
  "pattern": "^第[一二三四五六七八九十百千]+章",
  "file_path": "/path/to/file.docx",
  "num_script": 10
}
```

Response: `202 Accepted`
```json
{
  "success": true,
  "message": "task submitted",
  "task_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### Check task status

```
GET /api/v1/task/:id
```

Response:
```json
{
  "success": true,
  "task": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "running",
    "progress": 3,
    "total": 10,
    "save_dir": "./scripts",
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-01T00:05:00Z"
  }
}
```

Task status values: `pending`, `running`, `completed`, `failed`.

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8081` | Server port |
| `LLM_BASE_URL` | — | LLM API base URL (required) |
| `LLM_MODEL_NAME_LC` | — | Low-cost model name (required) |
| `LLM_MODEL_NAME_HC` | — | High-cost model name |
| `LLM_API_KEY` | — | LLM API key (required) |
| `MAX_CONCURRENCY` | `5` | Max concurrent LLM calls |
| `REQUEST_TIMEOUT_MS` | `300000` | Generation timeout in ms |
| `TEMP_DIR` | `./temp` | Upload temp directory |
| `SAVE_DIR` | `./scripts` | Script output directory |
| `MAX_UPLOAD_SIZE` | `104857600` | Max upload size in bytes |

## Docker

```bash
docker build -t doc2script .
docker run -p 8081:8081 \
  -e LLM_BASE_URL=http://your-llm-server/v1 \
  -e LLM_MODEL_NAME_LC=your-model \
  -e LLM_API_KEY=your-key \
  doc2script
```
