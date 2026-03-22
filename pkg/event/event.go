package event

type (
	Event struct {
		Payload any
		Headers map[string]interface{}
	}
)
