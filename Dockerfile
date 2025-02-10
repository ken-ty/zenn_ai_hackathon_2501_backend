FROM golang:1.21-alpine

# Cloud SDKの依存関係をインストール
RUN apk add --no-cache \
    python3 \
    py3-pip \
    curl \
    bash

# Cloud SDKのインストール
RUN curl -O https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-cli-458.0.0-linux-x86_64.tar.gz && \
    tar -xf google-cloud-cli-458.0.0-linux-x86_64.tar.gz && \
    ./google-cloud-sdk/install.sh --quiet && \
    rm google-cloud-cli-458.0.0-linux-x86_64.tar.gz

# PATHにCloud SDKを追加
ENV PATH $PATH:/google-cloud-sdk/bin

WORKDIR /app

# 依存関係のコピーとインストール
COPY go.mod go.sum ./
RUN go mod download

# ソースコードのコピー
COPY . .

# アプリケーションのビルド
RUN go build -o /server cmd/server/main.go

# サービスアカウントキーのコピー
COPY config/credentials/keyfile.json /keyfile.json
ENV GOOGLE_APPLICATION_CREDENTIALS=/keyfile.json

# ポートの設定
ENV PORT=8080
EXPOSE 8080

CMD ["/server"] 
