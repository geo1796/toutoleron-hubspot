package crm

import "fmt"

var cfg config

type config struct {
	initialized bool
	accountID   string
	baseURL     string
}

func (c config) objectsURL() string {
	return c.baseURL + "/objects"
}

func (c config) ownersURL() string {
	return c.baseURL + "/owners"
}

func Init(accountID, baseURL string) error {
	if cfg.initialized {
		return fmt.Errorf("crm already initialized")
	}
	cfg.initialized = true
	cfg.accountID = accountID
	cfg.baseURL = baseURL
	return nil
}
