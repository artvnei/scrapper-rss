package db

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

const (
	milvusAddr     = "91.107.187.118:19530"
	collectionName = "pure_news"
	vectorDim      = 768                                // must match the embedding model
	ollamaEndpoint = "https://ol-core.infiniteeight.io" // <‑‑ remote Ollama host
	ollamaModel    = "nomic-embed-text"                 // any *‑embed* model you pulled
)

type Data struct {
	ID       string    `json:"id"`
	URL      string    `json:"url"`
	Title    string    `json:"title"`
	Body     string    `json:"body"`
	Source   string    `json:"source"`
	Language string    `json:"language"`
	Vector   []float32 `json:"vector"`
}
type OllamaEmbedRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}
type OllamaEmbedResponse struct {
	Embeddings [][]float64 `json:"embeddings"` // or []float32 depending on your Ollama version
}

func SetVData(URL, Title, Body, Source, Language string) {
	client := resty.New()
	record_id := strconv.Itoa(int(time.Now().UTC().UnixNano()))
	// Sample data to insert
	data := Data{
		ID:       record_id,
		URL:      URL,
		Title:    Title,
		Body:     Body,
		Source:   Source,
		Language: Language,
		Vector:   []float32{0.1, 0.2, 0.3, 0.4},
	}

	// Send the POST request
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(data).
		Post("http://91.107.191.27:8080/v1/objects")

	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}

	// Print the response body
	//fmt.Println("Response Status:", resp.Status())
	fmt.Println("Response Body:", resp.String())
}
func getEmbeddingFromOllama(text string) ([]float64, error) {
	url := "http://46.4.72.240:11434/embed" // Ollama embedding API endpoint

	reqBody := OllamaEmbedRequest{
		Model: "nomic-embed-text", // Replace with your Ollama embedding model name
		Input: []string{text},
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var embedResp OllamaEmbedResponse
	err = json.Unmarshal(bodyBytes, &embedResp)
	if err != nil {
		return nil, err
	}

	if len(embedResp.Embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	return embedResp.Embeddings[0], nil
}
func SetSameDistance() {
	content := "HomeMARKETSEl Salvador To Buy One Bitcoin Every Day: President Bukele MARKETS El Salvador To Buy One Bitcoin Every Day: President Bukele By Shawn Amick November 17, 2022 Share FacebookTwitterLinkedinReddItEmailTelegramCopy URL Nayib Bukele, President of El Salvador, announced late last night that the country would be purchasing one bitcoin every day starting today. The move to dollar-cost-average (DCA) into bitcoin is common in the community, however novel for a nation state. Currently, the country holds a bitcoin treasury of 2,381 BTC, valued at over $39 million. Bukele has made a habit in the past of making large BTC purchases during times of market volatility and buying the dip. Outside of just purchasing BTC and holding it on balance for El Salvador, the Bukele administration has fostered the birth of events gathering world leaders from countries all over the world the learn about the financial freedom bitcoin adoption offers. In September, it was announced that over 30 countries with over 110 speakers, including Senator Indira Kempis from Mexico, would gather to discuss financial inclusion. During this visit, attendees were introduced to the financial applications of bitcoin and were able to see bitcoin in action at Bitcoin Beach. Then, in October, the State Treasurer from North Carolina in the U.S. traveled to El Salvador – on his own dime – to learn about the changes bitcoin has already made for the El Salvadoran economy. “What we witnessed in El Salvador is very useful in our efforts to encourage more support and understanding for digital assets and emerging technologies here in South Carolina,” said Dennis Fassuliotis, president of the South Carolina Emerging Technologies Association, at the time. As Bitcoin continues to foster throughout the El Salvadoran economy through new initiatives such as Bitcoin diplomas, Bukele and his administration clearly plan to double down on the country’s investment into a bitcoin-focused economy. It remains unclear how long the purchasing of 1 BTC per day will continue. Tagsbuy bitcoinDcael salvadorNewsPresident Bukele Share FacebookTwitterLinkedinReddItEmailTelegramCopy URL Previous articleScarce City Launches Bitcoin ECommerce Platform SatsCrapNext articleFTX Exchange Release Day One Bankruptcy Filing: “Complete Failure” Shawn AmickShawn Amick is an author \"Dick, Stan Greene\", podcast Co-host for Citizens of Blockchain, Substack blogger, and Bitcoin Magazine Contributor. RELATED ARTICLES MARKETS 00:12:51 The Enhanced Bitcoin Everything Indicator Unlocks Massive Profits May 9, 2025 MARKETS 00:11:41 Pro Tips For Maximizing MSTR Returns Using Bitcoin Market Data May 7, 2025 MARKETS Strategy’s Bitcoin Surge: Why MSTR Could Outperform BTC in 2025 May 2, 2025 Bitcoin BTC/USD $0.00 24hr %: 0.0% 24hr High: $0.00 24hr Low: $0.00 Error loading data. Check console for details. VIEW 150+ BITCOIN CHARTS LATEST NEWS Bitcoin Price Hits $104,000 As Demand Increases May 9, 2025 U.S. Vice President JD Vance To Speak At Bitcoin 2025 Conference May 9, 2025 New BIS Report Says Bitcoin Use Surges During Economic Stress May 9, 2025 Steak ‘n Shake Will Accept Bitcoin Payments in All U.S. Locations Starting Next Week May 9, 2025 Load more Get daily news in your inbox"
	embedding, err := getEmbeddingFromOllama(content)
	if err != nil {
		log.Fatalf("failed to get embedding: %v", err)
	}

	float32embedding := make([]float32, len(embedding))
	for i, v := range embedding {
		float32embedding[i] = float32(v)
	}
	fmt.Println("Successfully inserted document with embedding")
}

func MilvusSetData(title, content, source string, settime int64) {
	ctx := context.Background()

	// ─── 1. connect ────────────────────────────────────────────────────────────
	c, err := client.NewGrpcClient(ctx, milvusAddr)
	if err != nil {
		log.Fatalf("connect Milvus: %v", err)
	}
	defer c.Close()

	docstitle := []string{}
	docstitle = append(docstitle, title)

	docscontent := []string{}
	docscontent = append(docscontent, content)

	docssource := []string{}
	docssource = append(docssource, source)

	docssettime := []int64{}
	docssettime = append(docssettime, settime)

	embeddings := make([][]float32, 0, len(docstitle))
	for _, d := range docstitle {
		vec, err := embedWithOllama(d)
		if err != nil {
			log.Fatalf("embed: %v", err)
		}
		if len(vec) > 0 {
			embeddings = append(embeddings, vec)
		}
	}
	if len(embeddings) > 0 {
		titleCol := entity.NewColumnVarChar("title", docstitle)
		contentCol := entity.NewColumnVarChar("content", docstitle)
		identifyCol := entity.NewColumnVarChar("source", docssource)
		docssettimeCol := entity.NewColumnInt64("source", docssettime)
		vecCol := entity.NewColumnFloatVector("embedding", vectorDim, embeddings)

		c.Insert(ctx, collectionName, "", titleCol, contentCol, identifyCol, docssettimeCol, vecCol)
		//if _, err := c.Insert(ctx, collectionName, "", titleCol, contentCol, identifyCol, docssettimeCol, vecCol); err != nil {
		//	log.Fatalf("insert: %v", err)
		//}

		fmt.Println("✅  rows inserted ")
	}
}
func embedWithOllama(text string) ([]float32, error) {
	reqBody := map[string]interface{}{
		"model":  ollamaModel,
		"prompt": text,
	}
	body, _ := json.Marshal(reqBody)
	resp, err := http.Post(ollamaEndpoint+"/api/embeddings", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama error: %s", string(b))
	}
	var o struct {
		Embedding []float32 `json:"embedding"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&o); err != nil {
		return nil, err
	}
	return o.Embedding, nil
}
