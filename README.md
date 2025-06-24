# doc2script backend 

## fast start up

```bash
go run main.go
```

running on default port http://localhost:8080

## health check 
```
get /health
```

## upload file

```
POST /api/v1/upload
Content-Type: multipart/form-data

arguments:
- file: upload file
```
response
```json
{
  "success": true,
  "message": "文件上传成功",
  "files": ["example.txt"],
  "saved_path": "/path/to/saved/file"
}
```

## generate script

```
POST /api/v1/generate
Content-Type: application/json
{
  "model": "lc",                    // "hc" (高成本) 或 "lc" (低成本)
  "gentype": "d",                   // "d" (详细) 或 "r" (粗略)
  "pattern": "^第[一二三四五六七八九十百千]+章",  // 章节分割正则
  "file_path": "/path/to/file.txt", // 上传文件的路径
  "num_script": 10                  // 生成剧本数量
}
```