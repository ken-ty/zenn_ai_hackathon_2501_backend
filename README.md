# zenn_ai_hackathon_2501_backend

## 環境変数の設定

```bash
# プロジェクトの設定
export PROJECT_ID=zenn-ai-hackathon-2501
export LOCATION=us-central1
export REPO_NAME=zenn-ai-repo
export SERVICE_NAME=zenn-ai
export IMAGE_URL=${LOCATION}-docker.pkg.dev/${PROJECT_ID}/${REPO_NAME}/${SERVICE_NAME}:latest
```

## ローカル開発

### ビルド

```bash
pack build zenn_ai-linux --builder paketobuildpacks/builder-jammy-base --path . --platform linux/amd64
```

```bash
# ホットリロード環境のサーバー　（local専用
air
```

### 実行と確認

```bash
# ローカルで実行
docker run -p 8080:8080 zenn_ai-linux
# 動作確認
curl http://localhost:8080
```

## Cloud へのデプロイ

### 初期セットアップ（新規環境構築時のみ）

```bash
# Artifact Registryのリポジトリを作成
gcloud artifacts repositories create ${REPO_NAME} \
    --repository-format=docker \
    --location=${LOCATION}
# 認証の設定
gcloud auth configure-docker ${LOCATION}-docker.pkg.dev
```

### デプロイ手順

```bash
# ビルドとプッシュ
pack build --publish ${IMAGE_URL} --builder paketobuildpacks/builder-jammy-base --path . --platform linux/amd64
# Cloud Runへのデプロイ
gcloud run deploy ${SERVICE_NAME} --image ${IMAGE_URL} --platform managed --region ${LOCATION}  --allow-unauthenticated  --memory 512Mi --cpu 1 --port 8080
```

## Docker 操作

### 基本操作

```bash
# Dockerイメージの一覧を表示
docker images | grep zenn_ai
# 実行中のコンテナ確認
docker ps
# ログの確認
docker logs $(docker ps -q)
```

### トラブルシューティング

```bash
# Dockerイメージの削除（キャッシュクリア）
docker rmi zenn_ai-arm zenn_ai-linux
# コンテナの強制停止と削除
docker rm -f $(docker ps -aq)
```
