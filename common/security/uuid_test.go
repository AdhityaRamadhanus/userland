package security_test

import (
	"testing"

	"github.com/AdhityaRamadhanus/userland/security"
	"github.com/stretchr/testify/assert"
)

func TestGenerateUUID(t *testing.T) {
	sessionIDs := []string{}
	for i := 0; i < 1000; i++ {
		sessionID := security.GenerateUUID()
		sessionIDs = append(sessionIDs, sessionID)
	}

	// make sure id unique
	sessionIDTable := map[string]bool{}
	for _, sessionID := range sessionIDs {
		_, duplicate := sessionIDTable[sessionID]
		assert.False(t, duplicate)
		sessionIDTable[sessionID] = true
	}
}
