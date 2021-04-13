package servicenow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/byuoitav/auth/wso2"
)

type Client struct {
	Client *wso2.Client
}

type Incident struct {
	ID     string `json:"sys_id"`
	Number string `json:"number"`
}

func (c *Client) Incident(ctx context.Context, number string) (Incident, error) {
	url := "https://api.byu.edu:443/domains/servicenow/incident/v1.1/incident"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Incident{}, fmt.Errorf("unable to build request: %w", err)
	}

	query := req.URL.Query()
	query.Add("sysparm_limit", "1")
	query.Add("sysparm_query", fmt.Sprintf("number>=%s", number))
	req.URL.RawQuery = query.Encode()

	fmt.Printf("final request url: %q\n", req.URL.String())

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

	if len(wrapper.Result) == 0 {
		return Incident{}, errors.New("incident does not exist")
	}

	return wrapper.Result[0], nil
}

func (c *Client) AddInternalNote(ctx context.Context, number, note string) error {
	return nil
}
