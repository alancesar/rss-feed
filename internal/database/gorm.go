package database

import (
	"context"
	"rss-feed/internal/database/model"
	"rss-feed/pkg/rss"
	"time"

	"github.com/rs/zerolog"
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

func (g Gorm) SaveFeed(ctx context.Context, feed rss.Feed) error {
	zerolog.Ctx(ctx).Info().Str("feed", feed.Name).Int("articles", len(feed.Articles)).Msg("saving feed to database")
	m := model.NewFeedFromDomain(feed)
	if err := g.db.WithContext(ctx).Save(&m).Error; err != nil {
		return err
	}

	return nil
}

func (g Gorm) SaveArticle(ctx context.Context, feedID string, article rss.Article) error {
	zerolog.Ctx(ctx).Info().Str("feed_id", feedID).Str("article_title", article.Title).Msg("saving article to database")
	m := model.NewArticleFromDomain(article, feedID)
	if err := g.db.WithContext(ctx).Save(&m).Error; err != nil {
		return err
	}

	return nil
}

func (g Gorm) GetArticlesFromDate(ctx context.Context, date time.Time) ([]rss.Article, error) {
	var articles []model.Article
	if err := g.db.WithContext(ctx).
		Where("strftime('%Y-%m-%d', published_at) = ?", date.Format(time.DateOnly)).
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

func (g Gorm) GetAllFeedURLs(ctx context.Context) ([]string, error) {
	var feeds []model.Feed
	if err := g.db.WithContext(ctx).Find(&feeds).Error; err != nil {
		return nil, err
	}

	urls := make([]string, len(feeds))
	for i, feed := range feeds {
		urls[i] = feed.URL
	}

	zerolog.Ctx(ctx).Info().Int("count", len(urls)).Msg("retrieved feed URLs from database")
	return urls, nil
}

func (g Gorm) SaveImage(ctx context.Context, articleID string, img rss.Image) error {
	zerolog.Ctx(ctx).Info().Str("article_id", articleID).Str("path", img.Path).Msg("saving image to database")
	m := model.NewImageFromDomain(img, articleID)
	if err := g.db.WithContext(ctx).Save(&m).Error; err != nil {
		return err
	}

	return nil
}
