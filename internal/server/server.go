package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/zenn-dev/zenn-ai-hackathon/internal/models"
)

// QuizService はクイズ関連の操作を提供するインターフェース
type QuizService interface {
	CreateQuiz(ctx context.Context, imageData []byte, authorInterpretation string) (*models.Quiz, error)
	GetQuiz(ctx context.Context, quizID string) (*models.Quiz, error)
	GetRandomizedInterpretations(quiz *models.Quiz) []string
	VerifyAnswer(quiz *models.Quiz, selectedInterpretation string) bool
}

// Server はHTTPサーバーを表します
type Server struct {
	quizService QuizService
	mux         *http.ServeMux
}

// NewServer は新しいサーバーインスタンスを作成します
func NewServer(quizService QuizService) *Server {
	s := &Server{
		quizService: quizService,
		mux:         http.NewServeMux(),
	}
	s.routes()
	return s
}

// ServeHTTP はHTTPリクエストを処理します
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// routes はルーティングを設定します
func (s *Server) routes() {
	s.mux.HandleFunc("/upload", s.handleUpload)
	s.mux.HandleFunc("/quiz/", s.handleGetQuiz)
	s.mux.HandleFunc("/health", s.handleHealth)
}

// handleUpload は作品のアップロードを処理します
func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// ファイルの解析
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// 画像ファイルの取得
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 画像データの読み込み
	imageData := make([]byte, r.ContentLength)
	if _, err := file.Read(imageData); err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	// 解釈の取得
	interpretation := r.FormValue("interpretation")
	if interpretation == "" {
		http.Error(w, "Interpretation is required", http.StatusBadRequest)
		return
	}

	// クイズの作成
	quiz, err := s.quizService.CreateQuiz(r.Context(), imageData, interpretation)
	if err != nil {
		log.Printf("Failed to create quiz: %v", err)
		http.Error(w, "Failed to create quiz", http.StatusInternalServerError)
		return
	}

	// レスポンスの返却
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(quiz); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// handleGetQuiz はクイズの取得を処理します
func (s *Server) handleGetQuiz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// クイズIDの取得
	quizID := strings.TrimPrefix(r.URL.Path, "/quiz/")
	if quizID == "" {
		http.Error(w, "Quiz ID is required", http.StatusBadRequest)
		return
	}

	// クイズの取得
	quiz, err := s.quizService.GetQuiz(r.Context(), quizID)
	if err != nil {
		log.Printf("Failed to get quiz: %v", err)
		http.Error(w, "Failed to get quiz", http.StatusInternalServerError)
		return
	}

	// レスポンスの生成
	resp := models.NewQuizResponse(quiz, true)

	// レスポンスの返却
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// handleHealth はヘルスチェックを処理します
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write([]byte(`{"status":"ok"}`)); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}
