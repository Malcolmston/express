package jwtdecode

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
)

// ErrInvalidToken is returned when the token is malformed.
var ErrInvalidToken = errors.New("jwtdecode: invalid token")

func decodeSegment(seg string) (map[string]interface{}, error) {
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(seg, "="))
	if err != nil {
		return nil, ErrInvalidToken
	}
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, ErrInvalidToken
	}
	return m, nil
}

// Decode decodes the payload (second segment) of a JWT without verifying it.
func Decode(token string) (map[string]interface{}, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}
	return decodeSegment(parts[1])
}

// DecodeHeader decodes the header (first segment) of a JWT without verifying it.
func DecodeHeader(token string) (map[string]interface{}, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}
	return decodeSegment(parts[0])
}
