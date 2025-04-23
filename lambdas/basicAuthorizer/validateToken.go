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

	credentials := strings.Split(string(decoded), ":")
	if len(credentials) != 2 {
		return fmt.Errorf("token is not valid")
	}

	if credentials[0] != username {
		return fmt.Errorf("username is not correct")
	}

	if credentials[1] != password {
		return fmt.Errorf("password is not correct")
	}

	return nil
}
