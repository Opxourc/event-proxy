package endpoints

import "errors"

// Endpoint errors identify invalid endpoint input and registration failures.
var (
	ErrMissingEndpointURL        = errors.New("endpoint URL is required")
	ErrInvalidEndpointURL        = errors.New("endpoint URL must be a valid HTTP or HTTPS URL")
	ErrEndpointUnreachable       = errors.New("endpoint is not reachable")
	ErrEndpointAlreadyRegistered = errors.New("endpoint is already registered")
)

// Errors groups user-facing HTTP responses and internal log messages by channel.
var Errors = map[string]map[string]string{
	"HTTP": {
		"GetEndpointsFailure":       "Unable to retrieve endpoints.",
		"EncodeEndpointsFailure":    "Unable to encode endpoints.",
		"InvalidEndpointRequest":    "Request body must contain a valid endpoint URL.",
		"MissingEndpointURL":        "Endpoint URL is required.",
		"InvalidEndpointURL":        "Endpoint URL must be a valid HTTP or HTTPS URL.",
		"EndpointUnreachable":       "Endpoint could not be reached.",
		"EndpointAlreadyRegistered": "Endpoint is already registered.",
		"CreateEndpointFailure":     "Unable to register endpoint.",
		"DeleteEndpointFailure":     "Unable to remove endpoint.",
	},
	"Log": {
		"GetEndpointsFailure":       "failed to retrieve endpoints",
		"EncodeEndpointsFailure":    "failed to encode endpoints",
		"InvalidEndpointRequest":    "invalid endpoint request body",
		"MissingEndpointURL":        "endpoint URL is required",
		"InvalidEndpointURL":        "invalid endpoint URL",
		"EndpointUnreachable":       "endpoint availability check failed",
		"EndpointAlreadyRegistered": "endpoint is already registered",
		"CreateEndpointFailure":     "failed to register endpoint",
		"DeleteEndpointFailure":     "failed to remove endpoint",
	},
}
