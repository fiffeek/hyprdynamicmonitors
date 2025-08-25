package utils

import (
	"encoding/json"
	"errors"
	"fmt"
)

func UnmarshalResponse[T any](response []byte, v *T) error {
	if len(response) == 0 {
		return errors.New("empty json response")
	}

	err := json.Unmarshal(response, &v)
	if err != nil {
		return fmt.Errorf(
			"error while unmarshal: %w, response: %s",
			err,
			response,
		)
	}

	return nil
}
