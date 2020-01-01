// +build unit

package security_test

import (
	"testing"

	"github.com/AdhityaRamadhanus/userland/pkg/common/security"
)

func TestGenerateUUID_uniqueness(t *testing.T) {
	type args struct {
		count int
	}

	testCases := []struct {
		name string
		args args
	}{
		{
			name: "1000 unique UUID",
			args: args{
				count: 1000,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sessionIDs := []string{}
			for i := 0; i < tc.args.count; i++ {
				sessionID := security.GenerateUUID()
				sessionIDs = append(sessionIDs, sessionID)
			}

			// make sure id unique
			sessionIDTable := map[string]bool{}
			for _, sessionID := range sessionIDs {
				_, duplicate := sessionIDTable[sessionID]
				if duplicate {
					t.Fatalf("sessionID not unique %s", sessionID)
				}
				sessionIDTable[sessionID] = true
			}
		})
	}
}
