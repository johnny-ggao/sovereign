package crypto

import (
	"github.com/pquerna/otp/totp"
)

func GenerateTOTPSecret(issuer, email string) (string, string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: email,
	})
	if err != nil {
		return "", "", err
	}
	return key.Secret(), key.URL(), nil
}

func VerifyTOTP(secret, code string) (bool, error) {
	valid := totp.Validate(code, secret)
	return valid, nil
}
