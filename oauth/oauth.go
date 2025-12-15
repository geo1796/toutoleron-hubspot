package oauth

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

type OAuth interface {
	GetSetupURL() string

	GetTokensFromOAuthCode(code string) (refreshToken string, accessToken string, accessTokenExpiry time.Time, err error)
	GetTokensFromRefreshToken(rft string) (refreshToken string, accessToken string, accessTokenExpiry time.Time, err error)
	GetRefreshTokenInfo(rft string) (userEmail string, internalUserID string, err error)
	DeleteRefreshToken(rft string) error
}

type oauth struct {
	clientID     string
	clientSecret string
	redirectURI  string
	setupURL     string
	baseURL      string
}

type Config struct {
	ClientID     string
	ClientSecret string
	RedirectUri  string
	SetupURL     string
	BaseURL      string
}

func NewOAuth(cfg Config) OAuth {
	return &oauth{
		clientID:     cfg.ClientID,
		clientSecret: cfg.ClientSecret,
		redirectURI:  cfg.RedirectUri,
		setupURL:     cfg.SetupURL,
		baseURL:      cfg.BaseURL,
	}
}

func (o *oauth) GetSetupURL() string {
	return o.setupURL
}

func (o *oauth) GetTokensFromOAuthCode(code string) (refreshToken string, accessToken string, accessTokenExpiry time.Time, err error) {
	agent := fiber.Post(o.baseURL + "/token")

	formData := fiber.AcquireArgs()

	formData.Set("client_id", o.clientID)
	formData.Set("client_secret", o.clientSecret)
	formData.Set("redirect_uri", o.redirectURI)
	formData.Set("grant_type", "authorization_code")
	formData.Set("code", code)

	agent.Form(formData)

	resCode, body, errs := agent.Bytes()

	if len(errs) > 0 {
		return "", "", time.Time{}, fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("failed to retrieve auth from Hubspot, response code: %d, errs: %v", resCode, errors.Join(errs...)))
	}

	if resCode != fiber.StatusOK {
		return "", "", time.Time{}, fiber.NewError(resCode, fmt.Sprintf("failed to retrieve auth from Hubspot, response code: %d", resCode))
	}

	var bodyMap map[string]any
	if err := json.Unmarshal(body, &bodyMap); err != nil {
		return "", "", time.Time{}, err
	}

	var dto oauthTokens
	if err := json.Unmarshal(body, &dto); err != nil {
		return "", "", time.Time{}, err
	}

	return dto.RefreshToken, dto.AccessToken, dto.accessTokenExpiry(), nil
}

func (o *oauth) GetTokensFromRefreshToken(rft string) (refreshToken string, accessToken string, accessTokenExpiry time.Time, err error) {
	agent := fiber.Post(o.baseURL + "/token")

	agent.ContentType("application/x-www-form-urlencoded")

	reqBody := fmt.Sprintf("grant_type=refresh_token&refresh_token=%s&client_id=%s&client_secret=%s", rft, o.clientID, o.clientSecret)

	agent.Body([]byte(reqBody))

	resCode, body, errs := agent.Bytes()

	if len(errs) > 0 {
		return "", "", time.Time{}, fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("hubspot oauth failed to refresh access token, response code: %d, errs: %v", resCode, errors.Join(errs...)))
	}

	if resCode != fiber.StatusOK {
		return "", "", time.Time{}, fiber.NewError(resCode, fmt.Sprintf("hubspot oauth failed to refresh access token, response code: %d", resCode))
	}

	var dto oauthTokens

	if err := json.Unmarshal(body, &dto); err != nil {
		return "", "", time.Time{}, err
	}

	return dto.RefreshToken, dto.AccessToken, dto.accessTokenExpiry(), nil
}

func (o *oauth) GetRefreshTokenInfo(rft string) (userEmail string, internalUserID string, err error) {
	agent := fiber.Get(o.baseURL + "/refresh-tokens/" + rft)

	resCode, body, errs := agent.Bytes()

	if len(errs) > 0 {
		return "", "", fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("hubspot oauth failed to retrieve user from refresh token, response code: %d, errs: %v", resCode, errors.Join(errs...)))
	}

	if resCode != fiber.StatusOK {
		return "", "", fiber.NewError(resCode, fmt.Sprintf("hubspot oauth failed to retrieve user from refresh token, response code: %d", resCode))
	}

	var dto refreshTokenInfo

	if err := json.Unmarshal(body, &dto); err != nil {
		return "", "", err
	}

	return dto.UserEmail, strconv.Itoa(dto.InternalUserID), nil
}

func (o *oauth) DeleteRefreshToken(rft string) error {
	agent := fiber.Delete(o.baseURL + "/refresh-tokens/" + rft)

	resCode, _, errs := agent.Bytes()

	if len(errs) > 0 {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("hubspot oauth failed to delete refresh token: %v", errs))
	}

	if resCode != fiber.StatusNoContent {
		err := fiber.NewError(resCode, fmt.Sprintf("hubspot oauth failed to delete refresh token, response code: %d", resCode))
		return err
	}

	return nil
}

type oauthTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

func (dto oauthTokens) accessTokenExpiry() time.Time {
	return time.Now().Add(time.Second * time.Duration(dto.ExpiresIn-60))
}

type refreshTokenInfo struct {
	UserEmail      string `json:"user"`
	InternalUserID int    `json:"user_id"`
}
