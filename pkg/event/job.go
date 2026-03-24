package event

const CommandUpdateFeeds = "UPDATE_FEEDS"

type (
	Job struct {
		Command string `json:"command"`
	}
)

func NewJob(command string) Job {
	return Job{
		Command: command,
	}
}
