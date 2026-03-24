package event

const CommandUpdateFeeds = "UPDATE_FEEDS"

type (
	Job struct {
		Command string `json:"command"`
	}
)
