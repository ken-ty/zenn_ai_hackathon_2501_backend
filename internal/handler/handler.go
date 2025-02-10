package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"zenn_ai_hackathon_2501_backend/internal/ai"
	"zenn_ai_hackathon_2501_backend/internal/models"
	"zenn_ai_hackathon_2501_backend/internal/storage"
)

type Handler struct {
	storageClient *storage.Client
	aiClient      *ai.Client
}

func NewHandler(storageClient *storage.Client, aiClient *ai.Client) *Handler {
	return &Handler{
		storageClient: storageClient,
		aiClient:      aiClient,
	}
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
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
	if err := h.storageClient.UploadFile(r.Context(), filename, file); err != nil {
		http.Error(w, "Failed to upload file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// AI生成画像の作成（3枚）
	var fakeImages []string
	for i := 0; i < 3; i++ {
		file.Seek(0, 0)
		generatedImage, err := h.aiClient.GenerateImage(r.Context(), file)
		if err != nil {
			http.Error(w, "Failed to generate image: "+err.Error(), http.StatusInternalServerError)
			return
		}

		fakePath := fmt.Sprintf("generated/%s_fake%d%s", imageID, i, filepath.Ext(header.Filename))
		if err := h.storageClient.UploadFile(r.Context(), fakePath, generatedImage); err != nil {
			http.Error(w, "Failed to upload generated image: "+err.Error(), http.StatusInternalServerError)
			return
		}

		fakeImages = append(fakeImages, fakePath)
	}

	// クイズデータの取得と更新
	reader, err := h.storageClient.GetFile(r.Context(), "metadata/questions.json")
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

	if err := h.storageClient.UpdateQuestions(r.Context(), questions); err != nil {
		http.Error(w, "Failed to update questions: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// レスポンスの作成
	response := models.UploadResponse{
		ImageID:    imageID,
		StorageURL: fmt.Sprintf("gs://zenn-ai-hackathon-2501/%s", filename),
		Status:     "success",
	}

	// レスポンスをJSONとして送信
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Questions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	reader, err := h.storageClient.GetFile(r.Context(), "metadata/questions.json")
	if err != nil {
		http.Error(w, "Failed to read questions: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var response models.QuestionsResponse
	if err := json.NewDecoder(reader).Decode(&response); err != nil {
		http.Error(w, "Failed to decode questions: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 各画像に対して署名付きURLを生成
	for i := range response.Questions {
		signedURL, err := h.storageClient.GenerateSignedURL(r.Context(), response.Questions[i].OriginalImage)
		if err == nil {
			response.Questions[i].OriginalImage = signedURL
		}

		for j, fakePath := range response.Questions[i].FakeImages {
			signedURL, err := h.storageClient.GenerateSignedURL(r.Context(), fakePath)
			if err == nil {
				response.Questions[i].FakeImages[j] = signedURL
			}
		}
	}

	// レスポンスをJSONとして送信
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
