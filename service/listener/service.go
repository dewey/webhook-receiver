package listener

import (
	"github.com/go-kit/kit/log"
)

// Service is an interface for a incoming hook listener service
type Service interface {
	ValidToken(uuid string) (bool, error)
}

type service struct {
	l log.Logger
}

// NewService initializes a new usermanager service
func NewService(l log.Logger) *service {
	return &service{
		l: l,
	}
}

func (s *service) ValidToken(uuid string) (bool, error) {
	if uuid != "" && uuid == "a901cd2a-3ca0-416f-b1cd-51d67f440c18" {
		return true, nil
	}
	return false, nil
}
