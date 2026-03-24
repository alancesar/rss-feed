package event

type (
	Message struct {
		Payload any
		Headers map[string]interface{}
	}

	Delivery struct {
		Message
		Ack  func() error
		Nack func(requeue bool) error
	}
)
