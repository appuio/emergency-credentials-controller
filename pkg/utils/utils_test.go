package utils_test

import (
	"testing"

	"github.com/appuio/emergency-credentials-controller/pkg/utils"
	"github.com/stretchr/testify/require"
)

func Test_ParseJWTWithoutVerify(t *testing.T) {
	_, err := utils.ParseJWTWithoutVerify("invalid token")
	require.Error(t, err)

	v := "eyJhbGciOiJSUzI1NiIsImtpZCI6IkZqSlFfNTZqQ2l5M3JKSS1IRi1vU2czR2Y4cmtDQlN5OTh6QzdNRXV6Vm8ifQ.eyJhdWQiOlsiaHR0cHM6Ly9rdWJlcm5ldGVzLmRlZmF1bHQuc3ZjLmNsdXN0ZXIubG9jYWwiXSwiZXhwIjoxNjk4Njg4NjIwLCJpYXQiOjE2OTg2ODgwMjAsImlzcyI6Imh0dHBzOi8va3ViZXJuZXRlcy5kZWZhdWx0LnN2Yy5jbHVzdGVyLmxvY2FsIiwia3ViZXJuZXRlcy5pbyI6eyJuYW1lc3BhY2UiOiJkZWZhdWx0Iiwic2VydmljZWFjY291bnQiOnsibmFtZSI6ImVtZXJnZW5jeWFjY291bnQtc2FtcGxlIiwidWlkIjoiM2U1Njg1YWEtYmE2ZC00MGRjLTk4M2MtYzU3MWU5MDQ1YjM4In19LCJuYmYiOjE2OTg2ODgwMjAsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpkZWZhdWx0OmVtZXJnZW5jeWFjY291bnQtc2FtcGxlIn0.fE3w1IVhm6UZ7Il4ElNzuECChnyZVZ5Vug1DwaTHEJbDtT3Zf2vb-xYxC-PrE8oUrrdIfKEMsqC1xFszTeeQ-i_exFWkikN9_NRg38PuNg23segXC4I6lbBmO4fat2vyEWGQafI8lfnhyhli1ye2rslu1g86itLrDMO8ypykqkRsncoDJW22BVL-GoefjfQVQj-QmmFRzQSVdDYb3AVZAOmSzu_P_N-SjbYHbjrNOeWx-_8ftUS6SlbcVh3laL9MULpoz-gIRmDqs4MI1-CyJd4z9IDZJQI9tcyMsQYPgTYNW2TWxJHy2rxdoxQTR7LJ033tlVgh54jZRyybAaaD3w"
	_, err = utils.ParseJWTWithoutVerify(v)
	require.NoError(t, err)
}

func Test_SplitPublicKeyBlocks(t *testing.T) {
	input := `
leading garbage
-----BEGIN PGP PUBLIC KEY BLOCK-----
Public Key Block 1
-----END PGP PUBLIC KEY BLOCK-----
Some other content
-----BEGIN PGP PUBLIC KEY BLOCK-----
Public Key Block 2
-----END PGP PUBLIC KEY BLOCK-----`

	expected := []string{
		`-----BEGIN PGP PUBLIC KEY BLOCK-----
Public Key Block 1
-----END PGP PUBLIC KEY BLOCK-----`,
		`-----BEGIN PGP PUBLIC KEY BLOCK-----
Public Key Block 2
-----END PGP PUBLIC KEY BLOCK-----`,
	}

	result, err := utils.SplitPublicKeyBlocks(input)
	require.NoError(t, err)
	require.Equal(t, expected, result)
}

func Test_SplitPublicKeyBlocks_UnmatchedBegin(t *testing.T) {
	input := `
leading garbage
-----BEGIN PGP PUBLIC KEY BLOCK-----
Public Key Block 1
-----END PGP PUBLIC KEY BLOCK-----
Some other content
-----BEGIN PGP PUBLIC KEY BLOCK-----
asdasd
`

	expected := []string{
		`-----BEGIN PGP PUBLIC KEY BLOCK-----
Public Key Block 1
-----END PGP PUBLIC KEY BLOCK-----`,
	}

	result, err := utils.SplitPublicKeyBlocks(input)
	require.Error(t, err)
	require.Equal(t, expected, result)
}
