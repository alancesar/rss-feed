package database

import (
	"context"
	"errors"
	"rss-summary/internal/database/model"
	"rss-summary/pkg/rss"
	"time"

	"gorm.io/gorm"
)

type (
	Gorm struct {
		db *gorm.DB
	}
)

func NewGorm(db *gorm.DB) *Gorm {
	_ = db.AutoMigrate(&model.Article{}, &model.Feed{})
	return &Gorm{db: db}
}

func (g Gorm) SaveFeed(ctx context.Context, feed rss.Feed) error {
	m := model.NewFeedFromDomain(feed)
	if err := g.db.WithContext(ctx).Create(&m).Error; err != nil {
		return err
	}

	return nil
}

func (g Gorm) GetFeedByURL(ctx context.Context, url string) (rss.Feed, bool, error) {
	m := model.Feed{}
	if err := g.db.WithContext(ctx).
		Where("url = ?", url).
		First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return rss.Feed{}, false, nil
		}

		return rss.Feed{}, false, err
	}

	return m.ToDomain(), true, nil
}

func (g Gorm) GetArticlesFromDate(ctx context.Context, date time.Time) ([]rss.Article, error) {
	var articles []model.Article
	if err := g.db.WithContext(ctx).
		Where("strftime('%Y-%m-%d', published_at) = ?", date.Format(time.DateOnly)).
		Find(&articles).Error; err != nil {
		return nil, err
	}

	result := make([]rss.Article, len(articles))
	for i, a := range articles {
		result[i] = a.ToDomain()
	}
	return result, nil
}
