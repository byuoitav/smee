package opa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/byuoitav/auth/middleware"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Client struct {
	URL   string
	Token string
	Log   *zap.Logger
}

type opaResponse struct {
	DecisionID string    `json:"decision_id"`
	Result     opaResult `json:"result"`
}

type opaResult struct {
	Allow bool `json:"allow"`
}

type opaRequest struct {
	Input requestData `json:"input"`
}

type requestData struct {
	APIKey string `json:"api_key"`
	User   string `json:"user"`
	Path   string `json:"path"`
	Method string `json:"method"`
}

func (client *Client) Authorize() gin.HandlerFunc {
	return func(c *gin.Context) {

		// Initial data
		opaData := opaRequest{
			Input: requestData{
				Path:   c.FullPath(),
				Method: c.Request.Method,
			},
		}

		// use either the user netid for the authorization request or an
		// API key if one was used instead
		if user, ok := c.Request.Context().Value("user").(string); ok {
			opaData.Input.User = user
		} else if apiKey, ok := middleware.GetAVAPIKey(c.Request.Context()); ok {
			opaData.Input.APIKey = apiKey
		}

		// Prep the request
		oReq, err := json.Marshal(opaData)
		if err != nil {
			client.Log.Error(fmt.Sprintf("Error trying to create request to OPA: %s\n", err))
			if ginErr := c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("error while contacting authorization server")); ginErr != nil {
				client.Log.Error(fmt.Sprintf("Abort response: %s", ginErr))
			}
			return
		}

		req, _ := http.NewRequest(
			"POST",
			fmt.Sprintf("%s/v1/data/shipwright", client.URL),
			bytes.NewReader(oReq),
		)
		req.Header.Set("authorization", fmt.Sprintf("Bearer %s", client.Token))

		// Make the request
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			client.Log.Error(fmt.Sprintf("Error while making request to OPA: %s", err))
			if ginErr := c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("error while contacting authorization server")); ginErr != nil {
				client.Log.Error(fmt.Sprintf("Abort response: %s", ginErr))
			}
			return
		}
		if res.StatusCode != http.StatusOK {
			client.Log.Error(fmt.Sprintf("Got back non 200 status from OPA: %d", res.StatusCode))
			if ginErr := c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("error while contacting authorization server")); ginErr != nil {
				client.Log.Error(fmt.Sprintf("Abort response: %s", ginErr))
			}
			return
		}

		// Read the body
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			client.Log.Error(fmt.Sprintf("Unable to read body from OPA: %s", err))
			if ginErr := c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("error while contacting authorization server")); ginErr != nil {
				client.Log.Error(fmt.Sprintf("Abort response: %s", ginErr))
			}
			return
		}

		// Unmarshal the body
		oRes := opaResponse{}
		err = json.Unmarshal(body, &oRes)
		if err != nil {
			client.Log.Error(fmt.Sprintf("Unable to parse body from OPA: %s", err))
			if ginErr := c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("error while contacting authorization server")); ginErr != nil {
				client.Log.Error(fmt.Sprintf("Abort response: %s", ginErr))
			}
			return
		}

		// If OPA approved then allow the request, else reject with a 403
		if oRes.Result.Allow {
			c.Next()
		} else {
			if ginErr := c.AbortWithError(http.StatusForbidden, fmt.Errorf("Unauthorized")); ginErr != nil {
				client.Log.Error(fmt.Sprintf("Abort response: %s", ginErr))
			}
		}
	}
}
