package oauth

import "fmt"

var cfg Config

type Config struct {
	initialized  bool
	clientID     string
	clientSecret string
	redirectURL  string
	setupURL     string
}

func Init(
	clientID, clientSecret string,
	redirectURL, setupURL string,
) error {
	if cfg.initialized {
		return fmt.Errorf("oauth already initialized")
	}
	cfg.initialized = true
	cfg.clientID = clientID
	cfg.clientSecret = clientSecret
	cfg.redirectURL = redirectURL
	cfg.setupURL = setupURL
	return nil
}
