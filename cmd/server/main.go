package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"zenn_ai_hackathon_2501_backend/internal/ai"
	"zenn_ai_hackathon_2501_backend/internal/handler"
	"zenn_ai_hackathon_2501_backend/internal/storage"
)

func main() {
	ctx := context.Background()

	// Cloud Storageクライアントの初期化
	storageClient, err := storage.NewClient(ctx, "zenn-ai-hackathon-2501")
	if err != nil {
		log.Fatal(err)
	}

	// AI クライアントの初期化
	aiClient := ai.NewClient("zenn-ai-hackathon-2501", "us-central1")

	// ハンドラーの初期化
	h := handler.NewHandler(storageClient, aiClient)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// ルーティング設定
	http.HandleFunc("/health", h.HealthCheck)
	http.HandleFunc("/upload", h.Upload)
	http.HandleFunc("/questions", h.Questions)

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
