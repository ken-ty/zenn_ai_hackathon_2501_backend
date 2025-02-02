# zenn_ai_hackathon_2501_backend

## ビルド

```bash
# ビルド (Linux 向け)
pack build zenn_ai-linux --path . --builder=gcr.io/buildpacks/builder --platform linux/amd64
# ビルド (Apple Silicon 向け)
pack build zenn_ai-arm --builder paketobuildpacks/builder:base --path . --platform linux/arm64
```

## Docker 操作

```bash
# Dockerイメージの一覧を表示
docker images | grep zenn_ai-linux
# イメージを実行
docker run zenn_ai-linux
# イメージを削除
docker rmi zenn_ai-linux
# イメージの詳細情報を表示
docker inspect zenn_ai-linux
```
