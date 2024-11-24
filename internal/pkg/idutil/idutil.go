package idutil

import (
	"time"

	"math/rand"

	"github.com/oklog/ulid/v2"
)

func NewID() (string, error) {
	id, err := ulid.New(
		ulid.Timestamp(time.Now()),
		rand.New(rand.NewSource(time.Now().UnixNano())),
	)
	if err != nil {
		return "", err
	}
	return id.String(), nil
}
