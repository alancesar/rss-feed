package event

import "encoding/json"

type (
	Payload []byte

	Message struct {
		Topic   Topic
		Payload Payload
		Headers map[string]interface{}
	}

	Delivery struct {
		Message
		Ack  func()
		Nack func(requeue bool)
	}
)

func NewUpdateFeedCommand() Message {
	return Message{
		Topic:   TopicFeedJobs,
		Payload: NewPayload(Job{Command: CommandUpdateFeeds}),
	}
}

func NewImageFoundEvent(image Image) Message {
	return Message{
		Topic:   TopicFeedFound,
		Payload: NewPayload(image),
	}
}

func NewArticleFoundEvent(feed Feed) Message {
	return Message{
		Topic:   TopicFeedArticleFound,
		Payload: NewPayload(feed),
	}
}

func NewPayload(body any) Payload {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil
	}

	return jsonBody
}
