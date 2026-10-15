package utils

import "testing"

func TestHashPasswordAndCheckPasswordHash(t *testing.T) {
	const password = "correct-password"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() returned an error: %v", err)
	}
	if hash == password {
		t.Fatal("HashPassword() returned the original password")
	}
	if !CheckPasswordHash(password, hash) {
		t.Error("CheckPasswordHash() did not recognize the matching password")
	}
	if CheckPasswordHash("incorrect-password", hash) {
		t.Error("CheckPasswordHash() accepted an incorrect password")
	}
}

func TestCheckPasswordHashRejectsMalformedHash(t *testing.T) {
	if CheckPasswordHash("password", "not-a-bcrypt-hash") {
		t.Error("CheckPasswordHash() accepted a malformed hash")
	}
}
