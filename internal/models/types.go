package models

import "time"

// Quiz はクイズのデータモデルを表します
type Quiz struct {
	ID                   string    `json:"id"`
	ImagePath            string    `json:"image_path"`
	AuthorInterpretation string    `json:"author_interpretation"`
	AIInterpretation     string    `json:"ai_interpretation"`
	CreatedAt            time.Time `json:"created_at"`
}

// QuizList はクイズのリストを表します
type QuizList struct {
	Quizzes []*Quiz `json:"quizzes"`
}

// QuizListResponse はクイズ一覧のレスポンスを表します
type QuizListResponse struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}
