package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"testing"
)

func TestBcrypt(t *testing.T) {
	password, _ := bcrypt.GenerateFromPassword([]byte("a123456"), bcrypt.DefaultCost)
	fmt.Printf("%s", password)
}
