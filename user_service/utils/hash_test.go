package utils

import "testing"

func TestHashAndCheckPassword(t *testing.T) {
	password := "supersecret"

	hashed, err := HashPassword(password)
	if err != nil {
		t.Fatalf("hashing faild: %v", err)
	}

	if !CheckPasswordHash(password, hashed) {
		t.Errorf("password check failed")
	}

}
