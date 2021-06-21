package servicenow

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/byuoitav/auth/wso2"
	"go.uber.org/zap"
)

type Client struct {
	Client *wso2.Client
	Log    *zap.Logger
}

type Incident struct {
	ID               string `json:"sys_id,omitempty"`
	Number           string `json:"number,omitempty"`
	ShortDescription string `json:"short_description,omitempty"`
	Priority         string `json:"priority,omitempty"`
}

type IncidentRequest struct {
	ID               string `json:"sys_id,omitempty"`
	Number           string `json:"number,omitempty"`
	ShortDescription string `json:"short_description,omitempty"`
	WorkNotes        string `json:"work_notes,omitempty"`
	AssignmentGroup  string `json:"assignment_group,omitempty"`
	CallerNetID      string `json:"u_caller_net_id,omitempty"`
	Service          string `json:"business_service,omitempty"`
	Priority         string `json:"priority,omitempty"`
}

func (c *Client) Incident(ctx context.Context, id string) (Incident, error) {
	url := fmt.Sprintf("https://api.byu.edu:443/domains/servicenow/incident/v1.1/incident/%s", id)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Incident{}, fmt.Errorf("unable to build request: %w", err)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return Incident{}, fmt.Errorf("unable to do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return Incident{}, fmt.Errorf("%v response", resp.StatusCode)
	}

	var wrapper struct {
		Result Incident `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		return Incident{}, fmt.Errorf("unable to decode response: %w", err)
	}

	return wrapper.Result, nil
}

func (c *Client) IncidentByNumber(ctx context.Context, number string) (Incident, error) {
	if !strings.HasPrefix(number, "INC") {
		return Incident{}, errors.New("only INC tickets are supported")
	}

	url := "https://api.byu.edu:443/domains/servicenow/incident/v1.1/incident"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Incident{}, fmt.Errorf("unable to build request: %w", err)
	}

	query := req.URL.Query()
	query.Add("sysparm_limit", "2")
	query.Add("sysparm_query", fmt.Sprintf("number=%s", number))
	req.URL.RawQuery = query.Encode()

	resp, err := c.Client.Do(req)
	if err != nil {
		return Incident{}, fmt.Errorf("unable to do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return Incident{}, fmt.Errorf("%v response from servicenow incident api", resp.StatusCode)
	}

	var wrapper struct {
		Result []Incident `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		return Incident{}, fmt.Errorf("unable to decode response: %w", err)
	}

	switch len(wrapper.Result) {
	case 0:
		return Incident{}, errors.New("incident does not exist")
	case 1:
		return wrapper.Result[0], nil
	default:
		fmt.Printf("wrapper: %+v", wrapper)
		return Incident{}, errors.New("number matches multiple incidents")
	}
}

func (c *Client) AddInternalNote(ctx context.Context, id, note string) error {
	reqBody := IncidentRequest{
		WorkNotes: note,
	}

	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("unable to marshal body: %w", err)
	}

	url := fmt.Sprintf("https://api.byu.edu:443/domains/servicenow/incident/v1.1/incident/%s", id)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(reqBytes))
	if err != nil {
		return fmt.Errorf("unable to build request: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("%v response", resp.StatusCode)
	}

	return nil
}

func (c *Client) CreateIncident(ctx context.Context, inc IncidentRequest) (Incident, error) {
	reqBytes, err := json.Marshal(inc)
	if err != nil {
		return Incident{}, fmt.Errorf("unable to marshal body: %w", err)
	}

	url := "https://api.byu.edu:443/domains/servicenow/incident/v1.1/incident"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBytes))
	if err != nil {
		return Incident{}, fmt.Errorf("unable to build request: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return Incident{}, fmt.Errorf("unable to do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return Incident{}, fmt.Errorf("%v response", resp.StatusCode)
	}

	var wrapper struct {
		Result Incident `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		return Incident{}, fmt.Errorf("unable to decode response: %w", err)
	}

	return wrapper.Result, nil
}
