package servicenow

import (
	"context"
	"fmt"
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

	inc, err := client.IncidentByNumber(ctx, "INC0481325")
	is.NoErr(err)

	fmt.Printf("inc: %+v\n", inc)

	cmp, err := client.Incident(ctx, inc.ID)
	is.NoErr(err)
	is.Equal(inc, cmp)

	// add a note
	is.NoErr(client.AddInternalNote(ctx, inc.ID, "this is a test note from go test"))
}
