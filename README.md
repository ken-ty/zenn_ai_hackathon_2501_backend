# README

```bash
go run cmd/server/main.go
```

## 各エンドポイントを叩くcurl

```bash
curl http://localhost:8080/health
```

```bash
# zenn_ai_hackathon_2501_backend/ に移動
cd zenn_ai_hackathon_2501_backend/
# test.png をアップロード
curl -v -X POST -F "file=@test.png" http://localhost:8080/upload | jq .
```

```bash
curl http://localhost:8080/questions | jq .
```
