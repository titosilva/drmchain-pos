package uuid_test

import (
	"testing"

	"github.com/titosilva/drmchain-pos/internal/utils/uuid"
)

func Test__NewUuid__ShouldReturn__ValidUuidString(t *testing.T) {
	uuid := uuid.NewUuid()

	if len(uuid) == 0 {
		t.Error("Failed to generate UUID")
	}
}
