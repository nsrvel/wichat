package otp

import (
	"crypto/rand"
	"math/big"
	"strings"

	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
)

// Generate returns a random code of the given length from OTPAlphabet.
func Generate(length int) (string, error) {
	alphabet := constants.OTPAlphabet
	var out strings.Builder
	max := big.NewInt(int64(len(alphabet)))
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		out.WriteByte(alphabet[n.Int64()])
	}

	return out.String(), nil
}
