package event

type (
	Topic string
)

const (
	TopicFeedArticleFound      Topic = "rss.feed.article.found"
	TopicFeedArticleImageFound Topic = "rss.feed.article.image.found"
	TopicFeedJobs              Topic = "rss.feed.jobs"
)
