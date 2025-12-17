package crm

import (
	"errors"
)

var cfg config

type config struct {
	initialized bool
	accountID   string
	staticAuth  string
}

func Init(accountID, staticAuth string) error {
	if cfg.initialized {
		return errors.New("crm already initialized")
	}
	cfg.initialized = true
	cfg.accountID = accountID
	cfg.staticAuth = staticAuth
	return nil
}
