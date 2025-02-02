package models

import (
	"time"
)

// QuizResponse はクイズのレスポンス形式を定義します
type QuizResponse struct {
	ID                   string `json:"id"`
	ImageURL             string `json:"image_url"`
	CreatedAt            string `json:"created_at"`
	AuthorInterpretation string `json:"author_interpretation"`
	AIInterpretation     string `json:"ai_interpretation"`
}

// NewQuizResponse はQuizResponseを生成します
func NewQuizResponse(quiz *Quiz, imageURL string) *QuizResponse {
	return &QuizResponse{
		ID:                   quiz.ID,
		ImageURL:             imageURL,
		CreatedAt:            quiz.CreatedAt.Format(time.RFC3339),
		AuthorInterpretation: quiz.AuthorInterpretation,
		AIInterpretation:     quiz.AIInterpretation,
	}
}

// shuffle はスライスの要素をランダムに並び替えます
func shuffle(slice []string) {
	if len(slice) < 2 {
		return
	}
	for i := len(slice) - 1; i > 0; i-- {
		j := time.Now().UnixNano() % int64(i+1)
		slice[i], slice[j] = slice[j], slice[i]
	}
}
