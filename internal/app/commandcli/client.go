package commandcli

import (
	"context"
	"crypto/x509"
	"fmt"

	avcli "github.com/byuoitav/smee/proto"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Client struct {
	cli      avcli.AvCliClient
	cliToken string

	log *zap.Logger
}

func NewClient(ctx context.Context, cliAddr, cliToken string, logs *zap.Logger) (*Client, error) {
	pool, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("unable to get system cert pool: %s", err)
	}

	conn, err := grpc.DialContext(ctx, cliAddr, getTransportSecurityDialOption(pool))
	if err != nil {
		return nil, fmt.Errorf("unable to connect to avcli API: %s", err)
	}

	return &Client{
		cli:      avcli.NewAvCliClient(conn),
		cliToken: cliToken,
		log:      logs,
	}, nil
}
