package event

type (
	Message struct {
		Payload any
		Headers map[string]interface{}
	}

	Delivery struct {
		Payload []byte
		Headers map[string]interface{}
		Ack     func() error
		Nack    func(requeue bool) error
	}
)
