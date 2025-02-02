# zenn_ai_hackathon_2501_backend

## ビルド

```bash
pack build zenn_ai-linux --builder paketobuildpacks/builder-jammy-base --path . --platform linux/amd64
```

## Docker 操作

```bash
# Dockerイメージの一覧を表示
docker images | grep zenn_ai-linux
# イメージを実行
docker run -p 8080:8080 zenn_ai-linux
# 確認
curl http://localhost:8080
# イメージを削除
docker rmi zenn_ai-linux
# イメージの詳細情報を表示
docker inspect zenn_ai-linux
```
