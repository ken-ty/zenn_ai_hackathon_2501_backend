package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/zenn-dev/zenn-ai-hackathon/internal/models"
	"github.com/zenn-dev/zenn-ai-hackathon/internal/service"
)

type Handler struct {
	quizService service.QuizService
}

func NewHandler(quizService service.QuizService) *Handler {
	return &Handler{
		quizService: quizService,
	}
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "zenn-ai-hackathon-2501 is healthy")
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// ファイルの読み込み
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 投稿者の解釈を取得
	authorInterpretation := r.FormValue("interpretation")
	if authorInterpretation == "" {
		http.Error(w, "投稿者の解釈が必要です", http.StatusBadRequest)
		return
	}

	// ファイルデータの読み込み
	imageData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// クイズの作成
	quiz, err := h.quizService.CreateQuiz(r.Context(), imageData, authorInterpretation)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// レスポンスの作成
	imageURL, err := h.quizService.GetSignedImageURL(r.Context(), quiz.ImagePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := models.NewQuizResponse(quiz, imageURL)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) GetQuiz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// クイズIDの取得
	quizID := r.URL.Query().Get("id")
	if quizID == "" {
		http.Error(w, "クイズIDが必要です", http.StatusBadRequest)
		return
	}

	// クイズの取得
	quiz, err := h.quizService.GetQuiz(r.Context(), quizID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// レスポンスの作成
	imageURL, err := h.quizService.GetSignedImageURL(r.Context(), quiz.ImagePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := models.NewQuizResponse(quiz, imageURL)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
