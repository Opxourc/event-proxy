package events

// Errors groups user-facing HTTP responses and internal log messages by channel.
var Errors = map[string]map[string]string{
	"HTTP": {
		"InvalidRequestBody":    "Request body must contain valid JSON.",
		"EndpointNotRegistered": "Endpoint is not registered.",
		"InvalidMethod":         "Invalid HTTP method.",
		"InvalidURL":            "Endpoint URL must be a valid HTTP or HTTPS URL.",
		"MissingData":           "Event data is required.",
		"InvalidData":           "Event data must be valid JSON.",
		"DeliveryFailure":       "Unable to deliver event to endpoint.",
		"ProcessingFailure":     "Unable to process event.",
	},
	"Log": {
		"InvalidRequestBody":    "invalid event request body",
		"EndpointNotRegistered": "event target is not registered",
		"DeliveryFailure":       "event delivery failed",
		"ProcessingFailure":     "failed to process event",
		"WriteResponseFailure":  "failed to write upstream response",
	},
}
