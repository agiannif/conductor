// internal/handlers/helpers_test.go
package handlers_test

import "golang.org/x/crypto/bcrypt"

func bcryptHash(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(b), err
}
