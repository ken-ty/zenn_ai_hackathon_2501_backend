# ビルドステージ
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 依存関係のインストール
COPY go.mod go.sum ./
RUN go mod download

# ソースコードのコピー
COPY . .

# アプリケーションのビルド
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# 実行ステージ
FROM alpine:latest

WORKDIR /app

# 必要なディレクトリの作成
RUN mkdir -p /app/config/credentials

# 必要な証明書のインストール
RUN apk --no-cache add ca-certificates

# ビルドしたバイナリのコピー
COPY --from=builder /app/server .

# 実行ユーザーの設定
RUN adduser -D -g '' appuser && \
    chown -R appuser:appuser /app

USER appuser

# 環境変数の設定
ENV PORT=8080

# ヘルスチェックの設定
HEALTHCHECK --interval=5s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -q --spider http://localhost:${PORT}/health || exit 1

# ポートの公開
EXPOSE ${PORT}

# アプリケーションの実行
CMD ["./server"] 
