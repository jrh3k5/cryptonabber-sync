package http

import (
	"fmt"
	"io"
	"net/http"
)

// BuildUnexpectedStatusErr builds an error for an unexpected HTTP response.
func BuildUnexpectedStatusErr(response *http.Response) error {
	bodySampleLimit := int64(200)

	var bodyText string
	bodyBytes, bodyBytesErr := io.ReadAll(io.LimitReader(response.Body, bodySampleLimit))
	if bodyBytesErr != nil {
		bodyText = fmt.Sprintf("failed to read request body: %v", bodyBytesErr)
	} else {
		bodyText = string(bodyBytes)
	}

	return fmt.Errorf("unexpected response status code (%d); first %d bytes of body are: '%s'", response.StatusCode, bodySampleLimit, bodyText)
}
