package commandcli

import (
	"context"
	"crypto/x509"
	"fmt"
	"regexp"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type auth struct {
	token string
	user  string
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
		return grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	return grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(pool, ""))
}

func parseForCookie(id, list string) (string, error) {
	r, _ := regexp.Compile(`(?:\S*)`)
	cookies := r.FindAllString(list, -1)
	for _, c := range cookies {
		if strings.HasPrefix(c, id) {
			return strings.TrimPrefix(c, id+"="), nil
		}
	}
	return "", fmt.Errorf("cookie not found")
}

func getUserFromJWT(s string) (string, error) {
	parser := jwt.NewParser()
	token, _, err := parser.ParseUnverified(s, jwt.MapClaims{})
	if err != nil {
		return "", err
	}

	claims, _ := token.Claims.(jwt.MapClaims)
	if user, ok := claims["user"]; ok {
		return user.(string), nil
	}

	return "", fmt.Errorf("no user found in token")
}
