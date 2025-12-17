package crm

import (
	"errors"
)

var cfg config

type config struct {
	initialized bool
	accountID   string
	baseURL     string
	staticAuth  string
}

func (c config) objectsURL() string {
	return c.baseURL + "/objects"
}

func (c config) ownersURL() string {
	return c.baseURL + "/owners"
}

func Init(accountID, baseURL, staticAuth string) error {
	if cfg.initialized {
		return errors.New("crm already initialized")
	}
	cfg.initialized = true
	cfg.accountID = accountID
	cfg.baseURL = baseURL
	cfg.staticAuth = staticAuth
	return nil
}
