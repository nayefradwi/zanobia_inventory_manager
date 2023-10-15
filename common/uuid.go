package common

import (
	"log"
	"strings"

	"github.com/google/uuid"
)

func GenerateUuid() (string, error) {
	u, err := uuid.NewRandom()
	if err != nil {
		log.Printf("error generating uuid: %v", err)
		return "", err
	}
	uuidStr := strings.ReplaceAll(u.String(), "-", "")
	return uuidStr, nil
}
