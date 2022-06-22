package commandcli

import (
	"context"
	"crypto/x509"
	"encoding/json"

	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type auth struct {
	token string
	user  string
}

type JWTSegment struct {
	User string `json:"user"`
}

func (auth) RequireTransportSecurity() bool {
	return true
}

func (a auth) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "Bearer " + a.token,
		"x-user":        a.user,
	}, nil
}

func getTransportSecurityDialOption(pool *x509.CertPool) grpc.DialOption {
	if !(auth{}).RequireTransportSecurity() {
		return grpc.WithInsecure()
	}

	return grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(pool, ""))
}

func getUserFromJWT(s string) (string, error) {
	segBytes, err := jwt.DecodeSegment(s)
	if err != nil {
		return "", err
	}

	var seg JWTSegment
	err = json.Unmarshal(segBytes, &seg)
	if err != nil {
		return "", err
	}

	return seg.User, nil
}
