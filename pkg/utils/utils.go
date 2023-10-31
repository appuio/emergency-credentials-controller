package utils

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

// ParseJWTWithoutVerify parses a JWT token without verifying the signature.
// The token should be verified by using authentication.k8s.io/v1.TokenReview.
// https://support.hashicorp.com/hc/en-us/articles/18712750429843-How-to-check-validity-of-JWT-token-in-kubernetes
func ParseJWTWithoutVerify(token string, options ...jwt.ParserOption) (*jwt.Token, error) {
	t, err := jwt.Parse(token, nil, options...)
	if err != nil && errors.Is(err, jwt.ErrTokenUnverifiable) {
		return t, nil
	}
	return t, err
}
