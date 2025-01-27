package alertmanager

import "fmt"

var (
	ErrorInvalidJSONPayload = fmt.Errorf("invalid JSON payload")
	ErrorJSONDecoding       = fmt.Errorf("JSON decoding error")
	ErrorJSONEncoding       = fmt.Errorf("JSON encoding error")
	ErrorFailedToRetrieve   = fmt.Errorf("failed to retrieve alerts")
	ErrorSilenceIDNotFound  = fmt.Errorf("silence ID is required")
)
