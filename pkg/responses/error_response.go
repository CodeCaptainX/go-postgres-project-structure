package responses

import "fmt"

type ErrorResponse struct {
	MessageID string
	Err       error
}

func (e *ErrorResponse) ErrorString() string {
	return fmt.Sprintf("MessageID: %s, Error: %v",
		e.MessageID, e.Err)
}

func NewErrorResponse(messageId string, err error) *ErrorResponse {
	return &ErrorResponse{
		MessageID: messageId,
		Err:       err,
	}
}
