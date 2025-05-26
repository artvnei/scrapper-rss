package models

import (
	"time"

	"github.com/google/uuid"
)

// ========= scrap_news ==========

type News struct {
	ID                  uuid.UUID `json:"id"`
	URL                 string    `json:"url"`
	Title               string    `json:"title"`
	Body                string    `json:"body"`
	Source              string    `json:"source"`
	Language            string    `json:"language"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
	NewsHash            string    `json:"news_hash"`
	SemanticVector      string    `json:"semantic_vector"`
	Sentiment           string    `json:"sentiment"`
	SentimentScore      float64   `json:"sentiment_score"`
	Emotion             string    `json:"emotion"`
	Topic               string    `json:"topic"`
	ImportanceScore     int64     `json:"importance_score"`
	Summary             string    `json:"summary"`
	Status              string    `json:"status"`
	LanguagesTranslated string    `json:"languages_translated"`
	Phonemes            string    `json:"phonemes"`
	Image               string    `json:"image"`
	Score               float64   `json:"Score"`
	Prompt              string    `json:"prompt"`
	DeepseekPrompt      string    `json:"deepseek_prompt"`
	DeepseekWight       float64   `json:"deepseek_wight"`
	QwenPrompt          string    `json:"qwen_prompt"`
	QwenWight           float64   `json:"qwen_wight"`
	FalconPrompt        string    `json:"falcon_prompt"`
	FalconWight         float64   `json:"falcon_wight"`
	LlamaPrompt         string    `json:"llama_prompt"`
	LlamaWight          float64   `json:"llama_wight"`
	MistralPrompt       string    `json:"mistral_prompt"`
	MistralWight        float64   `json:"mistral_wight"`
	Summerize           string    `json:"summerize"`
	Sanatize            string    `json:"sanatize"`
	Phonotize           string    `json:"phonotize"`
	PrimaryTopic        string    `json:"primary_topic"`
	KeyEntities         string    `json:"key_entities"`
	ConfidenceScore     string    `json:"confidence_score"`
	BiasScore           string    `json:"bias_score"`
	Reasoning           string    `json:"reasoning"`
	Category            string    `json:"category"`
}
type RssFread struct {
	ID       uuid.UUID `json:"id"`
	Site     string    `json:"site"`
	Active   bool      `json:"active"`
	Language string    `json:"language"`
}
type Prompt struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	BlongsTo  string    `gorm:"column:blongs_to;index"`
	Status    string    `gorm:"index"`
	Prompt    string    `gorm:"type:text"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
