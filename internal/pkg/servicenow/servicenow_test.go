package servicenow

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/byuoitav/auth/wso2"
	"github.com/matryer/is"
)

func TestIncident(t *testing.T) {
	is := is.New(t)

	clientID := os.Getenv("SMEE_CLIENT_ID")
	clientSecret := os.Getenv("SMEE_CLIENT_SECRET")

	client := &Client{
		Client: wso2.New(clientID, clientSecret, "https://api.byu.edu", ""),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	inc, err := client.Incident(ctx, "")
	is.NoErr(err)
	is.True(inc.ID == "")
}
