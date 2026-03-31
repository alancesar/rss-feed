package event

type (
	Topic string
)

const (
	TopicFeedFound        Topic = "rss.feed.found"
	TopicFeedArticleFound Topic = "rss.feed.article.found"
	TopicFeedJobs         Topic = "rss.feed.jobs"
)
