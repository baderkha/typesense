package http

import "net/http"

// StatusIsSuccess : returns a boolean depedning on the status code success is >= 200 && <300
func StatusIsSuccess(statusCode int) bool {
	return statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices
}
