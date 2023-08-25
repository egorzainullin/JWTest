package encryptor

import (
	"golang.org/x/crypto/bcrypt"
)

func GetHash(toHash string) (hash string, err error) {
	toByte := []byte(toHash)
	bytes, err := bcrypt.GenerateFromPassword(toByte, 14)
	return string(bytes), err
}

func CheckHash(stringToCheck, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(stringToCheck))
	return err == nil
}
