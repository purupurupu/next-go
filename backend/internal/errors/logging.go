package errors

import (
	"github.com/rs/zerolog/log"
)

// InternalErrorWithLog logs the error with context and returns a generic internal error
// This is useful for logging the actual error details while returning a safe error to the client
func InternalErrorWithLog(err error, context string) *ApiError {
	log.Error().Err(err).Str("context", context).Msg("Internal error occurred")
	return InternalError()
}
