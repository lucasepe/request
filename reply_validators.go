package request

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// CheckStatus validates the response has an acceptable status code.
func CheckStatus(val any, acceptStatuses ...int) ReplyHandler {
	return func(res *http.Response) error {
		for _, code := range acceptStatuses {
			if res.StatusCode == code {
				return nil
			}
		}

		if res.Body == nil || val == nil {
			return fmt.Errorf("unexpected status: %d", res.StatusCode)
		}

		data, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		return json.Unmarshal(data, &val)
	}
}
