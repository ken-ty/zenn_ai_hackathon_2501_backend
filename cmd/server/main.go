package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"zenn_ai_hackathon_2501_backend/internal/models"
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
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// ファイルの取得
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// ファイル名の生成
	imageID := fmt.Sprintf("image_%d", time.Now().Unix())
	filename := fmt.Sprintf("original/%s%s", imageID, filepath.Ext(header.Filename))

	// Cloud Storageにアップロード
	if err := storageClient.UploadFile(r.Context(), filename, file); err != nil {
		http.Error(w, "Failed to upload file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// レスポンスの作成
	response := models.UploadResponse{
		ImageID:    imageID,
		StorageURL: fmt.Sprintf("gs://zenn-ai-hackathon-2501/%s", filename),
		Status:     "success",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func questionsHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: クイズ情報取得処理の実装
	w.WriteHeader(http.StatusNotImplemented)
}
