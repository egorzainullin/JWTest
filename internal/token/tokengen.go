package token

import (
	"encoding/base64"
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"testjwt.com/internal/config"
	"testjwt.com/internal/encryptor"
)

func GenTokens() (AccessAndRefreshTokens, error) {
	timeToAdd := time.Minute * 15
	expired := time.Now().Add(timeToAdd).Unix()
	preAccessToken := jwt.New(jwt.SigningMethodHS512)
	preAccessToken.Claims = &jwt.StandardClaims{ExpiresAt: expired}
	tokenStr, error := preAccessToken.SignedString([]byte(config.SecretKey))
	if error != nil {
		return AccessAndRefreshTokens{}, error
	}
	refreshTokenRand := uuid.New().String()
	refreshTokenbase64 := base64.StdEncoding.EncodeToString([]byte(refreshTokenRand))
	return AccessAndRefreshTokens{RefreshToken: refreshTokenbase64, AccessToken: tokenStr}, nil
}

func RegenTokens(encrypted string, tokens AccessAndRefreshTokens) (AccessAndRefreshTokens, error) {
	accessToken := tokens.AccessToken
	lastSix := accessToken[len(accessToken)-6:]
	stringFromTwoTokens := tokens.RefreshToken + lastSix
	if encryptor.CheckHash(stringFromTwoTokens, encrypted) {
		return GenTokens()
	}
	return AccessAndRefreshTokens{"", ""}, errors.New("Tokens are not correct")
}
