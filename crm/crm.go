package crm

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const baseURL = "https://api.hubapi.com/crm/v3/"
const ownersURL = baseURL + "owners/"
const objectsURL = baseURL + "objects/"

type CRM interface {
	FindObject(accessToken, objectTypeID, id, idProperty string, properties, associations []string) (Object, error)
	FindBatch(accessToken, objectTypeID string, ids, properties []string) ([]Object, error)
	FindObjectOwner(accessToken, objectOwnerID string) (*ObjectOwner, error)
	CreateObject(accessToken, objectTypeID string, properties map[string]any, associations []any) error
	UpdateObject(accessToken, objectTypeID, id, idProperty string, properties map[string]any) error
}

type crm struct{}

func New() (CRM, error) {
	if !cfg.initialized {
		return nil, fmt.Errorf("crm not initialized")
	}
	return &crm{}, nil
}

func setBearerToken(agent *fiber.Agent, accessToken string) error {
	if cfg.staticAuth != "" && accessToken == "" {
		agent.Request().Header.Add(fiber.HeaderAuthorization, "Bearer "+cfg.staticAuth)
		return nil
	}

	if cfg.staticAuth == "" && accessToken != "" {
		agent.Request().Header.Add(fiber.HeaderAuthorization, "Bearer "+accessToken)
		return nil
	}

	return errors.New("exactly one of staticAuth or accessToken must be set")
}

func (c *crm) FindObject(accessToken, objectTypeID, id, idProperty string, properties, associations []string) (Object, error) {
	endpoint := objectsURL + objectTypeID + "/" + id

	params := url.Values{}
	if idProperty != "" {
		params.Add("idProperty", idProperty)
	}
	if associations != nil {
		params.Add("associations", strings.Join(associations, ","))
	}
	if properties != nil {
		params.Add("properties", strings.Join(properties, ","))
	}
	if query := params.Encode(); query != "" {
		endpoint += "?" + query
	}

	agent := fiber.Get(endpoint)
	if err := setBearerToken(agent, accessToken); err != nil {
		return nil, err
	}

	resCode, resBody, errs := agent.Bytes()

	if len(errs) > 0 {
		return nil, fiber.NewError(fiber.StatusInternalServerError,
			fmt.Sprintf("hubspot crm request failed: {endpoint=%s, resCode=%d, resBody: %s, errs=%v}",
				endpoint, resCode, truncate(resBody, 512), errors.Join(errs...)))
	}

	if resCode != fiber.StatusOK {
		return nil, fiber.NewError(resCode, fmt.Sprintf("hubspot crm request failed {endpoint=%s, resCode=%d, resBody: %s}",
			endpoint, resCode, truncate(resBody, 512)))
	}

	var baseObject BaseObject

	if err := json.Unmarshal(resBody, &baseObject); err != nil {
		return nil, err
	}

	return &baseObject, nil
}

func (c *crm) FindBatch(accessToken, objectTypeID string, ids, properties []string) ([]Object, error) {
	if len(ids) == 0 {
		return nil, fmt.Errorf("objects must not be empty")
	}

	endpoint := objectsURL + objectTypeID + "/batch/read"

	type batchRequestDTO struct {
		Properties []string `json:"properties"`
		Inputs     []struct {
			ID string `json:"id"`
		} `json:"inputs"`
	}

	reqDTO := batchRequestDTO{Properties: properties}

	for _, id := range ids {
		reqDTO.Inputs = append(reqDTO.Inputs, struct {
			ID string `json:"id"`
		}{ID: id})
	}

	reqBody, err := json.Marshal(reqDTO)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal batch request body: %v", err)
	}

	agent := fiber.Post(endpoint)
	agent.Request().Header.Add(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	if err = setBearerToken(agent, accessToken); err != nil {
		return nil, err
	}

	agent.Body(reqBody)

	resCode, resBody, errs := agent.Bytes()

	if len(errs) > 0 {
		return nil, fiber.NewError(fiber.StatusInternalServerError,
			fmt.Sprintf("hubspot crm request failed {endpoint=%s, resCode=%d, resBody: %s, errs=%v}",
				endpoint, resCode, truncate(resBody, 512), errors.Join(errs...)))
	}

	if resCode != fiber.StatusOK {
		return nil, fiber.NewError(resCode, fmt.Sprintf("hubspot crm request failed {endpoint=%s, resCode=%d, resBody: %s}",
			endpoint, resCode, truncate(resBody, 512)))
	}

	type batchResponseDTO struct {
		Results []*BaseObject `json:"results"`
	}

	var resDTO batchResponseDTO
	if err = json.Unmarshal(resBody, &resDTO); err != nil {
		return nil, fmt.Errorf("failed to unmarshal batch response body: {endpoint: %s, resCode:%d, resBody: %s ,err: %v}",
			endpoint, resCode, truncate(resBody, 512), err)
	}

	if len(resDTO.Results) != len(ids) {
		return nil, fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("expected %d crm objects but got %d: {endpoint: %s, resCode: %d, resBody: %s}",
			len(ids), len(resDTO.Results), endpoint, resCode, truncate(resBody, 512)))
	}

	crmObjects := make([]Object, 0, len(resDTO.Results))

	for _, baseObject := range resDTO.Results {
		crmObjects = append(crmObjects, baseObject)
	}

	return crmObjects, nil
}

func (c *crm) FindObjectOwner(accessToken string, objectOwnerID string) (*ObjectOwner, error) {
	endpoint := ownersURL + objectOwnerID

	agent := fiber.Get(endpoint)
	if err := setBearerToken(agent, accessToken); err != nil {
		return nil, err
	}

	resCode, resBody, errs := agent.Bytes()

	if len(errs) > 0 {
		return nil, fiber.NewError(fiber.StatusInternalServerError,
			fmt.Sprintf("hubspot crm request failed {endpoint=%s, resCode=%d, resBody=%s, errs=%v}",
				endpoint, resCode, truncate(resBody, 512), errors.Join(errs...)))
	}

	if resCode != fiber.StatusOK {
		err := fiber.NewError(resCode, fmt.Sprintf("hubspot crm request failed {endpoint=%s, resCode=%d, resBody: %s}",
			endpoint, resCode, truncate(resBody, 512)))
		return nil, err
	}

	var objectOwner ObjectOwner

	if err := json.Unmarshal(resBody, &objectOwner); err != nil {
		return nil, fmt.Errorf("failed to unmarshal object owner response body: {endpoint: %s, resCode: %d, resBody: %s, err: %v}",
			endpoint, resCode, truncate(resBody, 512), err)
	}

	return &objectOwner, nil
}

func (c *crm) CreateObject(accessToken string, objectTypeID string, properties map[string]any, associations []any) error {
	endpoint := objectsURL + objectTypeID

	agent := fiber.Post(endpoint)
	agent.Request().Header.Add(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	if err := setBearerToken(agent, accessToken); err != nil {
		return err
	}

	reqBody, err := json.Marshal(map[string]any{
		"properties":   properties,
		"associations": associations,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal create object request body: {endpoint=%s, err=%v}", endpoint, err)
	}

	agent.Body(reqBody)

	resCode, resBody, errs := agent.Bytes()

	if len(errs) > 0 {
		return fiber.NewError(fiber.StatusInternalServerError,
			fmt.Sprintf("hubspot crm request failed {endpoint=%s, resCode=%d, resBody=%s, errs=%v}",
				endpoint, resCode, truncate(resBody, 512), errors.Join(errs...)))
	}

	if resCode != fiber.StatusCreated {
		return fiber.NewError(resCode, fmt.Sprintf("hubspot crm request failed {endpoint=%s, resCode=%d, resBody=%s}",
			endpoint, resCode, truncate(resBody, 512)))
	}

	return nil
}

func (c *crm) UpdateObject(accessToken, objectTypeID, id, idProperty string, properties map[string]any) error {
	endpoint := objectsURL + objectTypeID + "/" + id

	if idProperty != "" {
		params := url.Values{}
		params.Add("idProperty", idProperty)
		endpoint += "?" + params.Encode()
	}

	agent := fiber.Patch(endpoint)
	agent.Request().Header.Add(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	if err := setBearerToken(agent, accessToken); err != nil {
		return err
	}

	reqBody, err := json.Marshal(map[string]any{"properties": properties})
	if err != nil {
		return fmt.Errorf("failed to marshal update object request body: {endpoint=%s, err=%v}", endpoint, err)
	}

	agent.Body(reqBody)

	resCode, resBody, errs := agent.Bytes()

	if len(errs) > 0 {
		return fiber.NewError(fiber.StatusInternalServerError,
			fmt.Sprintf("hubspot crm request failed {endpoint=%s, resCode=%d, resBody=%s, errs=%v}",
				endpoint, resCode, truncate(resBody, 512), errors.Join(errs...)))
	}
	if resCode != fiber.StatusOK {
		return fiber.NewError(resCode, fmt.Sprintf("hubspot crm request failed {endpoint=%s, resCode=%d, resBody=%s}",
			endpoint, resCode, truncate(resBody, 512)))
	}

	return nil
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
