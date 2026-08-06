package application

import (
	"fmt"
	"github.com/google/uuid"
	"io"
)

func (s *Service) generateNonNilID(kind string) (uuid.UUID, error) {
	value, err := s.newID()
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: %s: %w", ErrGenerateID, kind, err)
	}
	if value == uuid.Nil {
		return uuid.Nil, fmt.Errorf("%w: %s: generated nil UUID", ErrGenerateID, kind)
	}
	return value, nil
}

func generateID() (uuid.UUID, error) { return uuid.NewRandom() }

func generateIDFrom(reader io.Reader) (uuid.UUID, error) { return uuid.NewRandomFromReader(reader) }
