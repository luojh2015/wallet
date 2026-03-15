package pwd

import (
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hash), nil
}

func CheckPasswordHash(password, hash string) bool {
	hashBytes, err := hex.DecodeString(hash)
	if err != nil {
		return false
	}

	err = bcrypt.CompareHashAndPassword(hashBytes, []byte(password))
	return err == nil
}
