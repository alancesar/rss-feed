package event

type (
	Topic string
)

const (
	TopicFeedArticleFound Topic = "rss.feed.article.found"
	TopicFeedFound        Topic = "rss.feed.found"
	TopicFeedJobs         Topic = "rss.feed.jobs"
)
