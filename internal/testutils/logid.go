package testutils

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/fiffeek/hyprdynamicmonitors/internal/utils"
	"github.com/stretchr/testify/assert"
)

type logline struct {
	LogID *utils.LogID `json:"log_id"`
}

func AssertLogsPresent(t *testing.T, logs []byte, expectedIDs []utils.LogID) {
	if len(expectedIDs) == 0 {
		return
	}

	text := string(logs)
	lines := strings.Split(text, "\n")
	seenIDs := []utils.LogID{}
	for _, line := range lines {
		var m logline
		err := json.Unmarshal([]byte(line), &m)
		if err != nil {
			continue
		}
		if m.LogID != nil {
			seenIDs = append(seenIDs, *m.LogID)
		}
	}
	assert.Equal(t, expectedIDs, seenIDs, "seen logs ids should match")
}
