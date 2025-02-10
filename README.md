# README

```bash
go run cmd/server/main.go
```

## 各エンドポイントを叩くcurl

### ローカル環境

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

### Cloud Run環境

```bash
# ヘルスチェック
curl https://backend-1074614507777.asia-northeast1.run.app/health

# 画像アップロード
curl -v -X POST -F "file=@test.png" https://backend-1074614507777.asia-northeast1.run.app/upload | jq .

# 質問一覧取得
curl https://backend-1074614507777.asia-northeast1.run.app/questions | jq .
```

## デプロイ

```bash
# イメージのビルドとプッシュ
gcloud builds submit --tag gcr.io/zenn-ai-hackathon-2501/backend

# Cloud Runへのデプロイ
gcloud run deploy backend \
  --image gcr.io/zenn-ai-hackathon-2501/backend \
  --platform managed \
  --region asia-northeast1 \
  --allow-unauthenticated \
  --service-account zenn-ai-backend@zenn-ai-hackathon-2501.iam.gserviceaccount.com
```
