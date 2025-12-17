package crm

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
)

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
	endpoint := cfg.objectsURL() + "/" + objectTypeID + "/" + id

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
		return nil, fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("hubspot crm request failed {endpoint=%s, code=%d, errs=%v}", endpoint, resCode, errors.Join(errs...)))
	}

	if resCode != fiber.StatusOK {
		return nil, fiber.NewError(resCode, fmt.Sprintf("hubspot crm request failed {endpoint=%s, code=%d}", endpoint, resCode))
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

	endpoint := cfg.objectsURL() + "/" + objectTypeID + "/batch/read"

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
		return nil, fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("hubspot crm request failed {endpoint=%s, code=%d, errs=%v}", endpoint, resCode, errors.Join(errs...)))
	}

	if resCode != fiber.StatusOK {
		return nil, fiber.NewError(resCode, fmt.Sprintf("hubspot crm request failed {endpoint=%s, code=%d}", endpoint, resCode))
	}

	type batchResponseDTO struct {
		Results []*BaseObject `json:"results"`
	}

	var resDTO batchResponseDTO
	if err = json.Unmarshal(resBody, &resDTO); err != nil {
		return nil, fmt.Errorf("failed to unmarshal batch response body: %v", err)
	}

	if len(resDTO.Results) != len(ids) {
		return nil, fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("expected %d crm objects but got %d", len(ids), len(resDTO.Results)))
	}

	crmObjects := make([]Object, 0, len(resDTO.Results))

	for _, baseObject := range resDTO.Results {
		crmObjects = append(crmObjects, baseObject)
	}

	return crmObjects, nil
}

func (c *crm) FindObjectOwner(accessToken string, objectOwnerID string) (*ObjectOwner, error) {
	endpoint := cfg.ownersURL() + "/" + objectOwnerID

	agent := fiber.Get(endpoint)
	if err := setBearerToken(agent, accessToken); err != nil {
		return nil, err
	}

	resCode, resBody, errs := agent.Bytes()

	if len(errs) > 0 {
		return nil, fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("hubspot crm request failed {endpoint=%s, code=%d, errs=%v}", endpoint, resCode, errors.Join(errs...)))
	}

	if resCode != fiber.StatusOK {
		err := fiber.NewError(resCode, fmt.Sprintf("hubspot crm request failed {endpoint=%s, code=%d}", endpoint, resCode))
		return nil, err
	}

	var objectOwner ObjectOwner

	if err := json.Unmarshal(resBody, &objectOwner); err != nil {
		return nil, fmt.Errorf("failed to unmarshal object owner response body: %v", err)
	}

	return &objectOwner, nil
}

func (c *crm) CreateObject(accessToken string, objectTypeID string, properties map[string]any, associations []any) error {
	endpoint := cfg.objectsURL() + "/" + objectTypeID

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
		return fmt.Errorf("failed to marshal create object request body: %v", err)
	}

	agent.Body(reqBody)

	resCode, _, errs := agent.Bytes()

	if len(errs) > 0 {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("hubspot crm request failed {endpoint=%s, code=%d, errs=%v}", endpoint, resCode, errors.Join(errs...)))
	}

	if resCode != fiber.StatusCreated {
		return fiber.NewError(resCode, fmt.Sprintf("hubspot crm request failed {endpoint=%s, code=%d}", endpoint, resCode))
	}

	return nil
}

func (c *crm) UpdateObject(accessToken, objectTypeID, id, idProperty string, properties map[string]any) error {
	endpoint := cfg.objectsURL() + "/" + objectTypeID + "/" + id

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
		return fmt.Errorf("failed to marshal update object request body: %v", err)
	}

	agent.Body(reqBody)

	resCode, _, errs := agent.Bytes()

	if len(errs) > 0 {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("hubspot crm request failed {endpoint=%s, code=%d, errs=%v}", endpoint, resCode, errors.Join(errs...)))
	}
	if resCode != fiber.StatusOK {
		return fiber.NewError(resCode, fmt.Sprintf("hubspot crm request failed {endpoint=%s, code=%d}", endpoint, resCode))
	}

	return nil
}
