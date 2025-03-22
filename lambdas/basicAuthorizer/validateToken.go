package main

import (
	"encoding/base64"
	"fmt"
	"strings"
)

// Validates base64 basic authorization token
func validateToken(token string) error {
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(token, "Basic "))
	if err != nil {
		return err
	}

	if password != strings.TrimSpace(string(decoded)) {
		return fmt.Errorf("password is not correct")
	}

	return nil
}
