package event

type (
	Topic string
)

const (
	TopicFeedFound             Topic = "rss.feed.found"
	TopicFeedArticleImageFound Topic = "rss.feed.article.image.found"
	TopicFeedJobs              Topic = "rss.feed.jobs"
)
