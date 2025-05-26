package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-shiori/go-readability"
	"github.com/google/uuid"
	"github.com/mmcdole/gofeed"
	"io"
	"log"
	"net/http"
	neturl "net/url"
	"rss-scraper/ai"
	"rss-scraper/db"
	"rss-scraper/models"
	"strings"
	"time"
)

// Article represents the extracted payload.
type ArticleExpand struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Image   string `json:"image"`
}
type Article struct {
	Feed      string `json:"feed"`
	Title     string `json:"title"`
	Link      string `json:"link"`
	Published string `json:"published"`
	Summary   string `json:"summary"`
	ScrapedAt string `json:"scraped_at"`
}

const userAgent = "Mozilla/5.0 (compatible; RSSScraper-Go/1.0; +https://example.com)"
const ua = "Mozilla/5.0 (compatible; ArticleExtractor/1.1; +https://example.com)"

func fetchFeed(url string, client *http.Client, parser *gofeed.Parser) ([]Article, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	feed, err := parser.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	var articles []Article
	for _, item := range feed.Items {
		art := Article{
			Feed:      url,
			Title:     item.Title,
			Link:      item.Link,
			Summary:   item.Description,
			ScrapedAt: time.Now().UTC().Format(time.RFC3339),
		}
		if item.PublishedParsed != nil {
			art.Published = item.PublishedParsed.Format(time.RFC3339)
		} else {
			art.Published = item.Published
		}
		articles = append(articles, art)
	}
	return articles, nil
}
func getImageURL(doc *goquery.Document) string {
	// Helper to get content attribute safely
	attr := func(s *goquery.Selection) string {
		if v, ok := s.Attr("content"); ok {
			return strings.TrimSpace(v)
		}
		return ""
	}

	// 1. Open Graph
	if og := attr(doc.Find(`meta[property="og:image"]`)); og != "" {
		return og
	}
	if og := attr(doc.Find(`meta[name="og:image"]`)); og != "" {
		return og
	}

	// 2. Twitter Card
	if tw := attr(doc.Find(`meta[property="twitter:image"]`)); tw != "" {
		return tw
	}
	if tw := attr(doc.Find(`meta[name="twitter:image"]`)); tw != "" {
		return tw
	}

	// 3. First img fallback
	if imgSrc, ok := doc.Find("img").First().Attr("src"); ok {
		return strings.TrimSpace(imgSrc)
	}

	return ""
}
func GetArticle(rawURL string) (ArticleExpand, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return ArticleExpand{}, err
	}
	req.Header.Set("User-Agent", ua)

	resp, err := client.Do(req)
	if err != nil {
		return ArticleExpand{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ArticleExpand{}, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ArticleExpand{}, err
	}

	// Parse base URL for readability
	base, err := neturl.Parse(rawURL)
	if err != nil {
		return ArticleExpand{}, err
	}

	// Extract main content via readability
	artDoc, err := readability.FromReader(bytes.NewReader(body), base)
	if err != nil {
		return ArticleExpand{}, err
	}

	// Extract best image via goquery
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return ArticleExpand{}, err
	}
	img := getImageURL(doc)

	return ArticleExpand{Title: artDoc.Title, Content: artDoc.TextContent, Image: img}, nil
}

func main() {

        db.InitDB()

	for {
		urls, err := db.GetActiveRss()
		if err != nil {
			log.Printf("failed to read feeds list: %v", err)
			time.Sleep(1 * time.Minute)
			continue
		}
		log.Printf("found %d feeds", len(urls))

		client := &http.Client{Timeout: 10 * time.Second}
		parser := gofeed.NewParser()

		for _, u := range urls {
			arts, err := fetchFeed(u.Site, client, parser)
			if err != nil {
				log.Printf("warn: could not fetch %s: %v", u.Site, err)
				continue
			}
			HandllerartDatat(arts)
		}

		log.Print("[crawler] cycle complete, restarting")
		time.Sleep(10 * time.Minute)
	}
}

func HandllerartDatat(arts []Article) {

	for _, artsData := range arts {

		var Time int64
		Time = 0
		Art, _ := GetArticle(artsData.Link)
		if len(artsData.Published) > 0 {
			t, err := time.Parse(time.RFC3339, artsData.Published)
			if err != nil {
				panic(err)
			}
			Time = t.Unix()
		}

		SummerizeGet := ai.SummerizeAction(Art.Content)

		cat := ai.CategoryAction(SummerizeGet)
		catjson := make(map[string]string)
		json.Unmarshal([]byte(cat), &catjson)

		time.Sleep(1 * time.Second)
		SanatizeGet := ai.SanatizeAction(SummerizeGet)
		PhonotizeGet := ai.PhonotizeAction(SummerizeGet)
		PrimaryTopicGet := ai.PrimaryTopicAction(Art.Content)
		KeyEntitiesGet := ai.KeyEntitiesAction(Art.Content)
		ConfidenceScoreGet := ai.ConfidenceScoreAction(Art.Content)
		BiasScoreGet := ai.BiasScoreAction(Art.Content)
		ReasoningGet := ai.ReasoningAction(Art.Content)

		item := &models.News{
			ID:              uuid.New(),
			URL:             artsData.Link,
			Title:           artsData.Title,
			Body:            Art.Content,
			Source:          artsData.Link,
			Image:           Art.Image,
			Language:        "en",
			NewsHash:        "",
			Summerize:       SummerizeGet,
			Sanatize:        SanatizeGet,
			Phonotize:       PhonotizeGet,
			PrimaryTopic:    PrimaryTopicGet,
			KeyEntities:     KeyEntitiesGet,
			ConfidenceScore: ConfidenceScoreGet,
			BiasScore:       BiasScoreGet,
			Reasoning:       ReasoningGet,
			Category:        catjson["category"],
		}

		log.Printf("[crawler] Attempting to save news: %s", item.Title)
		if err := db.SaveNews(item); err != nil {
			log.Printf("[crawler] ERROR saving news: %v", err)
		} else {
			log.Printf("[crawler] Successfully saved news: %s", item.Title)
		}

		db.MilvusSetData(artsData.Title, artsData.Summary, artsData.Link, Time)

		time.Sleep(1 * time.Second)
	}
}

func OldCat() {
	oldCat, _ := db.OldDataCat()
	for _, artsData := range oldCat {

		time.Sleep(1 * time.Second)
		cat := ai.CategoryAction(artsData.Summerize)
		catjson := make(map[string]string)
		json.Unmarshal([]byte(cat), &catjson)

		fmt.Println(catjson["category"])
		db.UpdateOldDataCat(catjson["category"], artsData.ID)
	}

}
