package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const ollamaHost = "https://localhost:11434/"

type Article struct {
	Title  string
	Text   string
	Source string
	Date   string // Format: "2006-01-02"
}
type OllamaGenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}
type OllamaResponse struct {
	Response string `json:"response"`
}
type OllamaEmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
}

func Check(title, text, source, date string) (float64, string) {
	article := Article{
		Title:  title,
		Text:   text,
		Source: source,
		Date:   date,
	}

	score, prompt := scoreArticle(article)

	return score, prompt
}
func scoreArticle(article Article) (float64, string) {
	sentiment, prompt := getSentimentScore(article.Text)
	credibility := getCredibilityScore(article.Text, article.Source)
	timeliness := getTimelinessScore(article.Date)
	relevance := getRelevanceScore(article.Text)

	// Weighted scores
	totalScore := 0.4*relevance + 0.3*credibility + 0.2*sentiment + 0.1*timeliness
	return totalScore, prompt
}
func ollamaGenerate(model string, prompt string) (string, error) {
	requestBody := OllamaGenerateRequest{
		Model:  model,
		Prompt: prompt,
		Stream: false,
	}

	jsonBody, _ := json.Marshal(requestBody)
	resp, err := http.Post(ollamaHost+"/api/generate", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var response OllamaResponse
	json.Unmarshal(body, &response)
	return response.Response, nil
}
func getSentimentScore(text string) (float64, string) {
	prompt := `You are a crypto-savvy NFT character with a bullish outlook on institutional adoption. Analyze the potential impact of a Bitcoin ETF approval on the market. Consider historical precedents, institutional interest, and potential price movements. Provide a balanced yet optimistic perspective.”
Character Opinion: ` + text

	response, err := ollamaGenerate("mistral", prompt)
	if err != nil {
		return 0.5, prompt // Fallback neutral score
	}

	switch response {
	case "POSITIVE":
		return 1.0, prompt
	case "NEGATIVE":
		return 0.0, prompt
	default:
		return 0.5, prompt
	}
}
func getCredibilityScore(text string, source string) float64 {
	prompt := fmt.Sprintf(`Evaluate the credibility of this news article from %s.
Consider factual accuracy and bias. Respond with a number between 0 and 1.

Article: %s
ONLY respond with the number:`, source, text)

	response, err := ollamaGenerate("llama2", prompt)
	if err != nil {
		return 0.5
	}

	var score float64
	_, err = fmt.Sscanf(response, "%f", &score)
	if err != nil {
		return 0.5
	}

	return score
}
func getTimelinessScore(dateStr string) float64 {
	articleDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return 0.5
	}

	daysOld := time.Since(articleDate).Hours() / 24
	if daysOld > 30 {
		return 0.0
	}
	return 1.0 - (daysOld / 30)
}
func getEmbedding(text string) ([]float64, error) {
	requestBody := map[string]interface{}{
		"model":  "mistral",
		"prompt": text,
	}

	jsonBody, _ := json.Marshal(requestBody)
	resp, err := http.Post(ollamaHost+"/api/embeddings", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var response OllamaEmbeddingResponse
	json.Unmarshal(body, &response)
	return response.Embedding, nil
}
func getRelevanceScore(text string) float64 {
	// Example: Compare with trending topics (you'll need to store these)
	trendingTopics := [][]float64{
		// Add pre-computed embeddings for trending topics
	}

	articleEmbedding, err := getEmbedding(text)
	if err != nil || len(articleEmbedding) == 0 {
		return 0.5
	}

	// Simple relevance calculation (max similarity)
	maxScore := 0.0
	for _, topic := range trendingTopics {
		score := cosineSimilarity(articleEmbedding, topic)
		if score > maxScore {
			maxScore = score
		}
	}
	return maxScore
}
func cosineSimilarity(a, b []float64) float64 {
	dotProduct := 0.0
	magnitudeA := 0.0
	magnitudeB := 0.0

	for i := range a {
		dotProduct += a[i] * b[i]
		magnitudeA += a[i] * a[i]
		magnitudeB += b[i] * b[i]
	}

	if magnitudeA == 0 || magnitudeB == 0 {
		return 0.0
	}
	return dotProduct / (magnitudeA * magnitudeB)
}
