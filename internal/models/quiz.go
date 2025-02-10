package models

import (
	"time"
)

// QuizResponse はクイズのAPIレスポンスを表します
type QuizResponse struct {
	ID              string   `json:"id"`
	ImagePath       string   `json:"image_path"`
	Interpretations []string `json:"interpretations"`
}

// NewQuizResponse は新しいQuizResponseを作成します
func NewQuizResponse(quiz *Quiz, randomized bool) *QuizResponse {
	resp := &QuizResponse{
		ID:        quiz.ID,
		ImagePath: quiz.ImagePath,
	}

	// 解釈をスライスに追加
	interpretations := []string{
		quiz.AuthorInterpretation,
		quiz.AIInterpretation,
	}

	// ランダム化が要求された場合は順序をシャッフル
	if randomized {
		shuffle(interpretations)
	}

	resp.Interpretations = interpretations
	return resp
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
