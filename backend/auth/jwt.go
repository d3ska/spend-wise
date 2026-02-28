package auth

import (
	"fmt"
	"strconv"
	"time"

	"backend/model"

	"github.com/golang-jwt/jwt/v5"
)

// jwtIssuer identifies this application's tokens, preventing tokens signed by
// other services with the same secret from being accepted.
const jwtIssuer = "spendwise"

// jwtAudience restricts tokens to the SpendWise API, providing an additional
// validation layer alongside issuer checks.
const jwtAudience = "spendwise-api"

// SignToken creates a signed JWT containing the user ID as the subject.
// Includes iss (issuer), aud (audience), and nbf (not before) claims to prevent token misuse.
func SignToken(userID model.UserID, secret string, expiry time.Duration) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Issuer:    jwtIssuer,
		Audience:  jwt.ClaimStrings{jwtAudience},
		Subject:   strconv.FormatInt(int64(userID), 10),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("signing token: %w", err)
	}
	return signed, nil
}

// VerifyToken parses and validates a JWT, returning the user ID from the subject claim.
// It enforces issuer and audience claims to ensure the token was minted by and intended
// for this application.
func VerifyToken(tokenStr string, secret string) (model.UserID, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(_ *jwt.Token) (any, error) {
		return []byte(secret), nil
	},
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithIssuer(jwtIssuer),
		jwt.WithAudience(jwtAudience),
	)
	if err != nil {
		return 0, fmt.Errorf("parsing token: %w", err)
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return 0, fmt.Errorf("invalid token claims")
	}

	id, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing subject: %w", err)
	}
	return model.UserID(id), nil
}
