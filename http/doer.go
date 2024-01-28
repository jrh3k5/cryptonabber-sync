package http

import "net/http"

// Doer is an abstract of the Do method on the provided HTTP client.
type Doer interface {
	// Do executes the given HTTP request.
	Do(request *http.Request) (*http.Response, error)
}
