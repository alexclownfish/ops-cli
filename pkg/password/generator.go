package password

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const (
	Uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	Lowercase = "abcdefghijklmnopqrstuvwxyz"
	Digits    = "0123456789"
	Special   = "#_-@~"
	AllChars  = Uppercase + Lowercase + Digits + Special
)

func Generate(length int) (string, error) {
	if length < 4 {
		return "", fmt.Errorf("密码长度至少为4")
	}
	
	for {
		password := make([]byte, length)
		for i := 0; i < length; i++ {
			n, err := rand.Int(rand.Reader, big.NewInt(int64(len(AllChars))))
			if err != nil {
				return "", err
			}
			password[i] = AllChars[n.Int64()]
		}
		
		if validate(string(password)) {
			return string(password), nil
		}
	}
}
func validate(password string) bool {
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false
	
	for _, c := range password {
		switch {
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= '0' && c <= '9':
			hasDigit = true
		case c == '#' || c == '_' || c == '-' || c == '@' || c == '~':
			hasSpecial = true
		}
	}
	
	return hasUpper && hasLower && hasDigit && hasSpecial
}
