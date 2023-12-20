package utils

import (
	"errors"
	"strings"

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

// SplitPublicKeyBlocks splits a string containing multiple PGP public key blocks into a slice of strings.
// Returns an error and the already found blocks if a block start is found without a matching block end.
func SplitPublicKeyBlocks(in string) ([]string, error) {
	var blocks []string
	const (
		BlockStart = "-----BEGIN PGP PUBLIC KEY BLOCK-----"
		BlockEnd   = "-----END PGP PUBLIC KEY BLOCK-----"
	)

	for {
		start := strings.Index(in, BlockStart)
		if start == -1 {
			break
		}
		end := strings.Index(in, BlockEnd)
		if end == -1 {
			return blocks, errors.New("unmatched PGP public key block start")
		}
		end += len(BlockEnd)

		blocks = append(blocks, in[start:end])
		in = in[end:]
	}

	return blocks, nil
}
