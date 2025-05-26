package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"rss-scraper/db"
	"time"
)

func PrimaryTopicAction(Text string) string {

	// ── Gather input text ───────────────────────────────────────────────────
	host := "https://ol-core.infiniteeight.io"
	model := "mistral"
	input := Text

	ctx := context.Background()
	Prompt, _ := db.GetPrompt(ctx, "primary_topic")
	prompt := fmt.Sprintf(
		"%s \n %s",
		Prompt,
		input,
	)

	reqBody := SummaryRequest{
		Model:  model,
		Prompt: prompt,
		Stream: false,
	}
	buf, err := json.Marshal(reqBody)
	if err != nil {
		log.Fatalf("marshal: %v", err)
	}

	url := host + "/api/generate"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(buf))
	httpReq.Header.Set("Content-Type", "application/json")
	httpClient := &http.Client{Timeout: 5 * time.Minute}
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		log.Fatalf("POST %s: %v", url, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("server returned %s:\n%s", resp.Status, string(body))
	}

	var out SummaryResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		log.Fatalf("decode: %v", err)
	}

	return out.Response
}
