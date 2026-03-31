package database

import (
	"context"
	"errors"
	"rss-feed/internal/database/model"
	"rss-feed/pkg/rss"
	"time"

	"gorm.io/gorm"
)

type (
	Gorm struct {
		db *gorm.DB
	}
)

func NewGorm(db *gorm.DB) *Gorm {
	_ = db.AutoMigrate(&model.Article{}, &model.Feed{}, &model.Image{})
	return &Gorm{db: db}
}

func (g Gorm) CreateFeed(ctx context.Context, feed rss.Feed) error {
	m := model.NewFeedFromDomain(feed)
	if err := g.db.WithContext(ctx).Create(&m).Error; err != nil {
		return err
	}

	return nil
}

func (g Gorm) GetFeedByURL(ctx context.Context, url string) (rss.Feed, bool, error) {
	var m model.Feed
	if err := g.db.WithContext(ctx).Where("url = ?", url).Preload("Articles").First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return rss.Feed{}, false, nil
		}

		return rss.Feed{}, false, err
	}

	return m.ToDomain(), true, nil
}

func (g Gorm) GetArticleByID(ctx context.Context, articleID string) (rss.Article, bool, error) {
	var m model.Article
	if err := g.db.WithContext(ctx).
		Where("id = ?", articleID).
		Preload("Feed").
		Preload("Images").
		First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return rss.Article{}, false, nil
		}

		return rss.Article{}, false, err
	}

	return m.ToDomain(), true, nil
}

func (g Gorm) GetArticlesFromDate(ctx context.Context, date time.Time) ([]rss.Article, error) {
	var articles []model.Article
	if err := g.db.WithContext(ctx).
		Where("strftime('%Y-%m-%d', published_at) = ?", date.Format(time.DateOnly)).
		Preload("Feed").
		Preload("Images").
		Find(&articles).Error; err != nil {
		return nil, err
	}

	result := make([]rss.Article, len(articles))
	for i, a := range articles {
		result[i] = a.ToDomain()
	}
	return result, nil
}

func (g Gorm) GetAllFeeds(ctx context.Context) ([]rss.Feed, error) {
	var feeds []model.Feed
	if err := g.db.WithContext(ctx).Find(&feeds).Error; err != nil {
		return nil, err
	}

	result := make([]rss.Feed, len(feeds))
	for i, f := range feeds {
		result[i] = f.ToDomain()
	}
	return result, nil
}

func (g Gorm) UpdateFeed(ctx context.Context, feed rss.Feed) error {
	m := model.NewFeedFromDomain(feed)
	return g.db.WithContext(ctx).Save(&m).Error
}

func (g Gorm) SaveImage(ctx context.Context, articleID string, img rss.Image) error {
	m := model.NewImageFromDomain(img, articleID)
	if err := g.db.WithContext(ctx).Save(&m).Error; err != nil {
		return err
	}

	return nil
}
