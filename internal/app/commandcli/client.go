package commandcli

import (
	"context"
	"crypto/x509"
	"fmt"
	"strings"

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

type response struct {
	Failed    []string
	Succeeded []string
}

func (resp *response) failed(status string) {
	resp.Failed = append(resp.Failed, status)
}

func (resp *response) successful(status string) {
	resp.Succeeded = append(resp.Succeeded, status)
}

func (resp *response) report() (report string) {
	if len(resp.Succeeded) > 0 {
		report += "Success: "
		for _, s := range resp.Succeeded {
			report += s + "; "
		}
		report = strings.TrimSuffix(report, "; ")

		report += "\n"
	}

	if len(resp.Failed) > 0 {
		report += "Failure: "
		for _, s := range resp.Failed {
			report += s + "; "
		}
		report = strings.TrimSuffix(report, "; ")
	}

	return
}
