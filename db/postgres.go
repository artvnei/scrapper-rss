package db

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"rss-scraper/models"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() error {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	var err error
	PgDataLoad := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		"91.107.187.118",
		"postgres",
		"bdypUz1l6LO3lGnErQ6C",
		"scraper",
		5432,
		"disable",
	)
	DB, err = gorm.Open(postgres.Open(PgDataLoad), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	return nil
}
func GetActiveRss() ([]models.RssFread, error) {
	var rssFread []models.RssFread
	err := DB.Where("active = ?", true).Order("RANDOM()").Find(&rssFread).Error
	return rssFread, err
}

func SaveNews(item *models.News) error {
	log.Printf("Attempting to save news with title: %s", item.Title)
	log.Printf("News hash: %s", item.NewsHash)
	var Err error
	// Set creation time
	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()
	tx := DB.Where("title = ?", item.Title).
		FirstOrCreate(&item, item)

	if tx.RowsAffected == 0 {
		Err = DB.Create(&item).Error
	}

	//if err := DB.Create(item).Error; err != nil {
	//	log.Printf("Error creating news: %v", err)
	//	return fmt.Errorf("failed to create news: %v", err)
	//}
	//log.Printf("Successfully created news with title: %s", item.Title)
	return Err
}
func GetPrompt(ctx context.Context, promptKey string) (string, error) {
	var p models.Prompt
	if err := DB.
		Where("blongs_to = ? AND status = ?", promptKey, "active").
		First(&p).Error; err != nil {
		return "", fmt.Errorf("db query: %w", err)
	}

	return p.Prompt, nil
}
func OldDataCat() ([]models.News, error) {
	var news []models.News
	err := DB.Find(&news).Error
	return news, err
}
func UpdateOldDataCat(cat string, id uuid.UUID) {
	updateAssetonsell := map[string]interface{}{
		"category": cat,
	}
	resultAction := DB.Table("news").Where("id = ? ", id).Updates(updateAssetonsell)
	if resultAction.Error != nil {
		panic(resultAction.Error)
	}
}

// GetAllRss returns all RSS feeds regardless of active status
func GetAllRss() ([]models.RssFread, error) {
	var feeds []models.RssFread
	err := DB.Order("site").Find(&feeds).Error
	return feeds, err
}

// SetRssActive toggles the active field of a feed
func SetRssActive(id uuid.UUID, active bool) error {
	return DB.Model(&models.RssFread{}).Where("id = ?", id).Update("active", active).Error
}

// GetRecentNews returns the latest news limited by the given number
func GetRecentNews(limit int) ([]models.News, error) {
	var news []models.News
	err := DB.Order("created_at desc").Limit(limit).Find(&news).Error
	return news, err
}

// CountNewsByStatus provides aggregated counts of news by status field
func CountNewsByStatus() (map[string]int64, error) {
	type res struct {
		Status string
		Count  int64
	}
	var rows []res
	if err := DB.Model(&models.News{}).Select("status, count(*) as count").Group("status").Scan(&rows).Error; err != nil {
		return nil, err
	}
	m := make(map[string]int64)
	for _, r := range rows {
		m[r.Status] = r.Count
	}
	return m, nil
}
