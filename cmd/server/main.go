package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"zenn_ai_hackathon_2501_backend/internal/storage"
)

var storageClient *storage.Client

func main() {
	ctx := context.Background()

	// Cloud Storageクライアントの初期化
	client, err := storage.NewClient(ctx, "zenn-ai-hackathon-2501")
	if err != nil {
		log.Fatal(err)
	}
	storageClient = client

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// ルーティング設定
	http.HandleFunc("/health", healthCheckHandler)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/questions", questionsHandler)

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: 画像アップロード処理の実装
	w.WriteHeader(http.StatusNotImplemented)
}

func questionsHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: クイズ情報取得処理の実装
	w.WriteHeader(http.StatusNotImplemented)
}
