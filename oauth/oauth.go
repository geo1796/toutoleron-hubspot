package oauth

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

const baseURL = "https://api.hubapi.com/oauth/v1/"

type OAuth interface {
	GetSetupURL() string

	GetTokensFromOAuthCode(code string) (refreshToken string, accessToken string, accessTokenExpiry time.Time, err error)
	GetTokensFromRefreshToken(rft string) (refreshToken string, accessToken string, accessTokenExpiry time.Time, err error)
	GetRefreshTokenInfo(rft string) (userEmail string, internalUserID string, err error)
	DeleteRefreshToken(rft string) error
}

type oauth struct{}

func New() (OAuth, error) {
	if !cfg.initialized {
		return nil, fmt.Errorf("oauth not initialized")
	}
	return &oauth{}, nil
}

func (o *oauth) GetSetupURL() string {
	return cfg.setupURL
}

func (o *oauth) GetTokensFromOAuthCode(code string) (refreshToken string, accessToken string, accessTokenExpiry time.Time, err error) {
	endpoint := baseURL + "token"
	agent := fiber.Post(endpoint)

	formData := fiber.AcquireArgs()

	formData.Set("client_id", cfg.clientID)
	formData.Set("client_secret", cfg.clientSecret)
	formData.Set("redirect_uri", cfg.redirectURL)
	formData.Set("grant_type", "authorization_code")
	formData.Set("code", code)

	agent.Form(formData)

	resCode, resBody, errs := agent.Bytes()

	if len(errs) > 0 {
		return "", "", time.Time{}, fiber.NewError(fiber.StatusInternalServerError,
			fmt.Sprintf("failed to retrieve auth from Hubspot, {endpoint: %s, resCode: %d, resBody: %s, errs: %v}",
				endpoint, resCode, truncate(resBody, 512), errors.Join(errs...)))
	}

	if resCode != fiber.StatusOK {
		return "", "", time.Time{}, fiber.NewError(resCode,
			fmt.Sprintf("failed to retrieve auth from Hubspot: {endpoint: %s, resCode: %d, resBody: %s}",
				endpoint, resCode, truncate(resBody, 512)))
	}

	var bodyMap map[string]any
	if err := json.Unmarshal(resBody, &bodyMap); err != nil {
		return "", "", time.Time{}, err
	}

	var dto oauthTokens
	if err := json.Unmarshal(resBody, &dto); err != nil {
		return "", "", time.Time{}, err
	}

	return dto.RefreshToken, dto.AccessToken, dto.accessTokenExpiry(), nil
}

func (o *oauth) GetTokensFromRefreshToken(rft string) (refreshToken string, accessToken string, accessTokenExpiry time.Time, err error) {
	endpoint := baseURL + "token"
	agent := fiber.Post(endpoint)

	agent.ContentType("application/x-www-form-urlencoded")

	reqBody := fmt.Sprintf("grant_type=refresh_token&refresh_token=%s&client_id=%s&client_secret=%s", rft, cfg.clientID, cfg.clientSecret)

	agent.Body([]byte(reqBody))

	resCode, resBody, errs := agent.Bytes()

	if len(errs) > 0 {
		return "", "", time.Time{}, fiber.NewError(fiber.StatusInternalServerError,
			fmt.Sprintf("hubspot oauth failed to refresh access token: {endpoint: %s, resCode: %d, resBody: %s, errs: %v}",
				endpoint, resCode, truncate(resBody, 512), errors.Join(errs...)))
	}

	if resCode != fiber.StatusOK {
		return "", "", time.Time{}, fiber.NewError(resCode,
			fmt.Sprintf("hubspot oauth failed to refresh access token: {endpoint: %s, resCode: %d, resBody: %s}",
				endpoint, resCode, truncate(resBody, 512)))
	}

	var dto oauthTokens

	if err := json.Unmarshal(resBody, &dto); err != nil {
		return "", "", time.Time{}, err
	}

	return dto.RefreshToken, dto.AccessToken, dto.accessTokenExpiry(), nil
}

func (o *oauth) GetRefreshTokenInfo(rft string) (userEmail string, internalUserID string, err error) {
	endpoint := baseURL + "refresh-tokens/" + rft
	agent := fiber.Get(endpoint)

	resCode, resBody, errs := agent.Bytes()

	if len(errs) > 0 {
		return "", "", fiber.NewError(fiber.StatusInternalServerError,
			fmt.Sprintf("hubspot oauth failed to retrieve user from refresh token: {endpoint: %s, resCode: %d, resBody: %s,  errs: %v}",
				endpoint, resCode, truncate(resBody, 512), errors.Join(errs...)))
	}

	if resCode != fiber.StatusOK {
		return "", "", fiber.NewError(resCode,
			fmt.Sprintf("hubspot oauth failed to retrieve user from refresh token: {endpoint: %s, resCode: %d, resBody: %s}",
				endpoint, resCode, truncate(resBody, 512)))
	}

	var dto refreshTokenInfo

	if err := json.Unmarshal(resBody, &dto); err != nil {
		return "", "", err
	}

	return dto.UserEmail, strconv.Itoa(dto.InternalUserID), nil
}

func (o *oauth) DeleteRefreshToken(rft string) error {
	endpoint := baseURL + "refresh-tokens/" + rft
	agent := fiber.Delete(endpoint)

	resCode, resBody, errs := agent.Bytes()

	if len(errs) > 0 {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf(
			"hubspot oauth failed to delete refresh token: {endpoint: %s, resCode: %d, resBody: %s, errs: %v}",
			endpoint, resCode, truncate(resBody, 512), errors.Join(errs...)))
	}

	if resCode != fiber.StatusNoContent {
		err := fiber.NewError(resCode, fmt.Sprintf(
			"hubspot oauth failed to delete refresh token: {endpoint: %s, resCode: %d, resBody: %s}",
			endpoint, resCode, truncate(resBody, 512)))
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

// Truncate returns a UTF-8 string representation of b, limited to maxBytes.
// If truncation happens, it appends a short suffix indicating how many bytes were omitted.
func truncate(b []byte, maxBytes int) string {
	if maxBytes <= 0 {
		return ""
	}
	if len(b) <= maxBytes {
		return string(b)
	}
	omitted := len(b) - maxBytes
	return fmt.Sprintf("%s...[truncated %d bytes]", string(b[:maxBytes]), omitted)
}
