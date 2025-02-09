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

	"zenn_ai_hackathon_2501_backend/internal/ai"
	"zenn_ai_hackathon_2501_backend/internal/models"
	"zenn_ai_hackathon_2501_backend/internal/storage"
)

var (
	storageClient *storage.Client
	aiClient      *ai.Client
)

func main() {
	ctx := context.Background()

	// Cloud Storageクライアントの初期化
	client, err := storage.NewClient(ctx, "zenn-ai-hackathon-2501")
	if err != nil {
		log.Fatal(err)
	}
	storageClient = client

	// AI クライアントの初期化
	aiClient = ai.NewClient("zenn-ai-hackathon-2501", "us-central1")

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

	// AI生成画像の作成（3枚）
	var fakeImages []string
	for i := 0; i < 3; i++ {
		// ファイルポインタを先頭に戻す
		file.Seek(0, 0)

		// AI生成画像の生成
		generatedImage, err := aiClient.GenerateImage(r.Context(), file)
		if err != nil {
			http.Error(w, "Failed to generate image: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// 生成画像の保存
		fakePath := fmt.Sprintf("generated/%s_fake%d%s", imageID, i, filepath.Ext(header.Filename))
		if err := storageClient.UploadFile(r.Context(), fakePath, generatedImage); err != nil {
			http.Error(w, "Failed to upload generated image: "+err.Error(), http.StatusInternalServerError)
			return
		}

		fakeImages = append(fakeImages, fakePath)
	}

	// クイズデータの取得
	reader, err := storageClient.GetFile(r.Context(), "metadata/questions.json")
	if err != nil {
		http.Error(w, "Failed to read questions: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var questions models.QuestionsResponse
	if err := json.NewDecoder(reader).Decode(&questions); err != nil {
		http.Error(w, "Failed to decode questions: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 新しいクイズの追加
	newQuestion := models.Question{
		ID:            imageID,
		OriginalImage: filename,
		FakeImages:    fakeImages,
		CorrectIndex:  0,
		CreatedAt:     time.Now().UTC().Format(time.RFC3339),
	}
	questions.Questions = append(questions.Questions, newQuestion)

	// メタデータの更新
	if err := storageClient.UpdateQuestions(r.Context(), questions); err != nil {
		http.Error(w, "Failed to update questions: "+err.Error(), http.StatusInternalServerError)
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
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// メタデータファイルの読み込み
	reader, err := storageClient.GetFile(r.Context(), "metadata/questions.json")
	if err != nil {
		http.Error(w, "Failed to read questions: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var response models.QuestionsResponse
	if err := json.NewDecoder(reader).Decode(&response); err != nil {
		http.Error(w, "Failed to decode questions: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
