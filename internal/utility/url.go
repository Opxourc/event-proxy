package utility

import (
	"context"
	"fmt"
	"net/http"
)

// PingURL sends a HEAD request to determine whether an HTTP endpoint responds
// successfully.
func PingURL(ctx context.Context, rawURL string, client *http.Client) error {
	requestCtx, cancel := context.WithTimeout(ctx, client.Timeout)
	defer cancel()

	request, err := http.NewRequestWithContext(requestCtx, http.MethodHead, rawURL, nil)
	if err != nil {
		return err
	}

	response, err := client.Do(request)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("endpoint returned HTTP status %d", response.StatusCode)
	}
	return nil
}
