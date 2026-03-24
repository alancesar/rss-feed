package event

type (
	Event struct {
		Payload any
		Headers map[string]interface{}
	}

	Message struct {
		Payload []byte
		Headers map[string]interface{}
		Ack     func() error
		Nack    func(requeue bool) error
	}
)
