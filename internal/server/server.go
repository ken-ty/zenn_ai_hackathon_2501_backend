package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/zenn-dev/zenn-ai-hackathon/internal/logging"
	"github.com/zenn-dev/zenn-ai-hackathon/internal/service"
)

// Server はHTTPサーバーを表します
type Server struct {
	quizService service.QuizService
	mux         *http.ServeMux
}

// enableCORS はCORSを有効にするミドルウェアです
func enableCORS(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CORSヘッダーの設定
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type, Accept-Ranges, Content-Range")
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		// プリフライトリクエストの処理
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Max-Age", "86400") // プリフライトリクエストのキャッシュ時間（24時間）
			w.WriteHeader(http.StatusOK)
			return
		}

		handler.ServeHTTP(w, r)
	})
}

// NewServer は新しいサーバーを作成します
func NewServer(quizService service.QuizService) *Server {
	s := &Server{
		quizService: quizService,
		mux:         http.NewServeMux(),
	}
	s.setupRoutes()
	return s
}

// ServeHTTP はHTTPリクエストを処理します
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// CORSミドルウェアを適用
	handler := enableCORS(s.mux)
	handler.ServeHTTP(w, r)
}

// setupRoutes はルーティングを設定します
func (s *Server) setupRoutes() {
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/quizzes", s.handleGetQuizList)
	s.mux.HandleFunc("/quizzes/", s.handleGetQuiz)
	s.mux.HandleFunc("/upload", s.handleUpload)
	s.mux.HandleFunc("/verify-answer", s.handleVerifyAnswer)
	s.mux.HandleFunc("/delete-all-quizzes", s.handleDeleteAllQuizzes)
	logging.Info("routes: ルーティングを設定しました")
}

// handleUpload は画像とその解釈をアップロードするハンドラーです
func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	logging.Info("handleUpload: リクエストを受信")

	// マルチパートフォームの解析
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		logging.Error("handleUpload: フォームの解析に失敗: %v", err)
		http.Error(w, "フォームの解析に失敗しました", http.StatusBadRequest)
		return
	}

	// 画像ファイルの取得
	file, header, err := r.FormFile("file")
	if err != nil {
		logging.Error("handleUpload: 画像ファイルの取得に失敗: %v", err)
		http.Error(w, "画像ファイルの取得に失敗しました", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 投稿者の解釈の取得
	interpretation := r.FormValue("interpretation")
	if interpretation == "" {
		logging.Error("handleUpload: 投稿者の解釈が空です")
		http.Error(w, "投稿者の解釈が必要です", http.StatusBadRequest)
		return
	}

	// 画像の検証と保存
	validator := service.NewImageValidator(5 * 1024 * 1024) // 5MB
	buf, err := validator.ValidateAndCopy(file, header.Filename)
	if err != nil {
		logging.Error("handleUpload: 画像の検証に失敗: %v", err)
		http.Error(w, fmt.Sprintf("画像の検証に失敗しました: %v", err), http.StatusBadRequest)
		return
	}

	// クイズの作成
	quiz, err := s.quizService.CreateQuiz(r.Context(), buf.Bytes(), interpretation)
	if err != nil {
		logging.Error("handleUpload: クイズの作成に失敗: %v", err)
		http.Error(w, fmt.Sprintf("クイズの作成に失敗しました: %v", err), http.StatusInternalServerError)
		return
	}

	// レスポンスの送信
	w.Header().Set("Content-Type", "application/json")
	response := struct {
		ID                   string `json:"id"`
		ImageURL             string `json:"image_url"`
		CreatedAt            string `json:"created_at"`
		AuthorInterpretation string `json:"author_interpretation"`
		AIInterpretation     string `json:"ai_interpretation"`
	}{
		ID:                   quiz.ID,
		CreatedAt:            quiz.CreatedAt.Format("2006-01-02 15:04:05"),
		AuthorInterpretation: quiz.AuthorInterpretation,
		AIInterpretation:     quiz.AIInterpretation,
	}

	// 画像URLの生成
	imageURL, err := s.quizService.GetSignedImageURL(r.Context(), quiz.ImagePath)
	if err != nil {
		logging.Error("handleUpload: 画像URLの生成に失敗: %v", err)
		http.Error(w, fmt.Sprintf("画像URLの生成に失敗しました: %v", err), http.StatusInternalServerError)
		return
	}
	response.ImageURL = imageURL

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logging.Error("handleUpload: レスポンスの送信に失敗: %v", err)
		http.Error(w, "レスポンスの送信に失敗しました", http.StatusInternalServerError)
		return
	}

	logging.Info("handleUpload: クイズの作成に成功: id=%s", quiz.ID)
}

// handleGetQuiz はクイズを取得するハンドラーです
func (s *Server) handleGetQuiz(w http.ResponseWriter, r *http.Request) {
	quizID := strings.TrimPrefix(r.URL.Path, "/quizzes/")
	logging.Info("handleGetQuiz: クイズID=%s の取得を開始", quizID)

	// クイズの取得
	quiz, err := s.quizService.GetQuiz(r.Context(), quizID)
	if err != nil {
		logging.Error("handleGetQuiz: クイズの取得に失敗: %v", err)
		http.Error(w, fmt.Sprintf("クイズの取得に失敗しました: %v", err), http.StatusNotFound)
		return
	}
	logging.Debug("handleGetQuiz: クイズ取得成功: imagePath=%s", quiz.ImagePath)

	// 画像URLの生成
	imageURL, err := s.quizService.GetSignedImageURL(r.Context(), quiz.ImagePath)
	if err != nil {
		logging.Error("handleGetQuiz: 画像URLの生成に失敗: %v", err)
		http.Error(w, fmt.Sprintf("画像URLの生成に失敗しました: %v", err), http.StatusInternalServerError)
		return
	}
	logging.Debug("handleGetQuiz: 画像URL生成成功: URL=%s", imageURL)

	// レスポンスの構築
	response := struct {
		ID                   string `json:"id"`
		ImageURL             string `json:"image_url"`
		CreatedAt            string `json:"created_at"`
		AuthorInterpretation string `json:"author_interpretation"`
		AIInterpretation     string `json:"ai_interpretation"`
	}{
		ID:                   quiz.ID,
		ImageURL:             imageURL,
		CreatedAt:            quiz.CreatedAt.Format("2006-01-02 15:04:05"),
		AuthorInterpretation: quiz.AuthorInterpretation,
		AIInterpretation:     quiz.AIInterpretation,
	}

	// レスポンスの送信
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logging.Error("handleGetQuiz: レスポンスの送信に失敗: %v", err)
		http.Error(w, "レスポンスの送信に失敗しました", http.StatusInternalServerError)
		return
	}

	logging.Info("handleGetQuiz: レスポンス送信完了: quizID=%s", quizID)
}

// handleHealth はヘルスチェックを処理します
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// handleGetQuizList はクイズ一覧を取得するハンドラーです
func (s *Server) handleGetQuizList(w http.ResponseWriter, r *http.Request) {
	logging.Info("handleGetQuizList: リクエストを受信")

	// キャッシュ制御の設定
	w.Header().Set("Cache-Control", "max-age=15")

	// クイズ一覧の取得
	quizzes, err := s.quizService.GetQuizList(r.Context())
	if err != nil {
		logging.Error("handleGetQuizList: クイズ一覧の取得に失敗: %v", err)
		http.Error(w, fmt.Sprintf("クイズ一覧の取得に失敗しました: %v", err), http.StatusInternalServerError)
		return
	}

	// レスポンスの構築
	response := make([]struct {
		ID        string `json:"id"`
		CreatedAt string `json:"created_at"`
	}, len(quizzes))

	for i, quiz := range quizzes {
		response[i] = struct {
			ID        string `json:"id"`
			CreatedAt string `json:"created_at"`
		}{
			ID:        quiz.ID,
			CreatedAt: quiz.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	// レスポンスの送信
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logging.Error("handleGetQuizList: レスポンスの送信に失敗: %v", err)
		http.Error(w, "レスポンスの送信に失敗しました", http.StatusInternalServerError)
		return
	}

	logging.Info("handleGetQuizList: クイズ一覧の送信完了: count=%d", len(quizzes))
}

// handleVerifyAnswer は解答を検証するハンドラーです
func (s *Server) handleVerifyAnswer(w http.ResponseWriter, r *http.Request) {
	logging.Info("handleVerifyAnswer: リクエストを受信")

	// リクエストの解析
	var request struct {
		QuizID                 string `json:"quiz_id"`
		SelectedInterpretation string `json:"selected_interpretation"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		logging.Error("handleVerifyAnswer: リクエストの解析に失敗: %v", err)
		http.Error(w, "リクエストの解析に失敗しました", http.StatusBadRequest)
		return
	}

	// クイズの取得
	quiz, err := s.quizService.GetQuiz(r.Context(), request.QuizID)
	if err != nil {
		logging.Error("handleVerifyAnswer: クイズの取得に失敗: %v", err)
		http.Error(w, fmt.Sprintf("クイズの取得に失敗しました: %v", err), http.StatusNotFound)
		return
	}

	// 解答の検証
	isCorrect := s.quizService.VerifyAnswer(quiz, request.SelectedInterpretation)

	// レスポンスの送信
	response := struct {
		IsCorrect bool `json:"is_correct"`
	}{
		IsCorrect: isCorrect,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logging.Error("handleVerifyAnswer: レスポンスの送信に失敗: %v", err)
		http.Error(w, "レスポンスの送信に失敗しました", http.StatusInternalServerError)
		return
	}

	logging.Info("handleVerifyAnswer: 解答検証完了: quizID=%s, isCorrect=%v", request.QuizID, isCorrect)
}

// handleDeleteAllQuizzes は全てのクイズを削除するハンドラーです
func (s *Server) handleDeleteAllQuizzes(w http.ResponseWriter, r *http.Request) {
	logging.Info("handleDeleteAllQuizzes: リクエストを受信")

	// DELETEメソッドのみ許可
	if r.Method != http.MethodDelete {
		logging.Error("handleDeleteAllQuizzes: 不正なメソッド: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 全クイズの削除
	if err := s.quizService.DeleteAllQuizzes(r.Context()); err != nil {
		logging.Error("handleDeleteAllQuizzes: クイズの削除に失敗: %v", err)
		http.Error(w, fmt.Sprintf("クイズの削除に失敗しました: %v", err), http.StatusInternalServerError)
		return
	}

	// レスポンスの送信
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "全てのクイズを削除しました",
	})

	logging.Info("handleDeleteAllQuizzes: 全クイズの削除に成功")
}
