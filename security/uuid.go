package security

import (
	"strings"

	uuid "github.com/satori/go.uuid"
)

func GenerateUUID() string {
	id := uuid.NewV4()
	splitID := strings.Split(id.String(), "-")
	return strings.Join(splitID, "")
}
