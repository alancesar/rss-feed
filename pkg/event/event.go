package event

import "encoding/json"

type (
	Payload []byte

	Message struct {
		Payload Payload
		Headers map[string]interface{}
	}

	Delivery struct {
		Message
		Ack  func() error
		Nack func(requeue bool) error
	}
)

func NewPayload(body any) Payload {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil
	}

	return jsonBody
}
