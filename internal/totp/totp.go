package totp

import (
	"github.com/pquerna/otp/totp"
)

type TotpService struct {
	issuer string
}

func NewTotpService(issuer string) *TotpService {
	return &TotpService{issuer: issuer}
}

func (s *TotpService) GenerateSecret(email string) (string, string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      s.issuer,
		AccountName: email,
	})
	if err != nil {
		return "", "", err
	}
	return key.Secret(), key.URL(), nil
}

func (s *TotpService) ValidateCode(secret, code string) bool {
	return totp.Validate(code, secret)
}
