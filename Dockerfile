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

# ビルドしたバイナリのコピー
COPY --from=builder /app/server .

# 必要な証明書のインストール
RUN apk --no-cache add ca-certificates

# 実行ユーザーの設定
RUN adduser -D -g '' appuser
USER appuser

# ポートの公開
EXPOSE 8080

# アプリケーションの実行
CMD ["./server"] 
