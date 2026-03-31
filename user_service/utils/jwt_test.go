package utils

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMain(m *testing.M) {
	_ = godotenv.Load("../.env")

	// Ensure JWT_SECRET is set for tests (fallback when .env is absent, e.g. in CI)
	if os.Getenv("JWT_SECRET") == "" {
		os.Setenv("JWT_SECRET", "test-secret-key-for-ci")
	}

	os.Exit(m.Run())
}

func TestGenerateJwt(t *testing.T) {

	userId := "67e6c37b452365a9c0e36eae"
	role := "user"
	oid, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		t.Error("error", err.Error())
	}

	token, err := GenerateJWT(oid, role)
	if err != nil {
		t.Error("error", err.Error())
	}

	claims, err := ValidateJWT(token)

	if err != nil {
		t.Error("error", err.Error())
	}

	if claims["user_id"] != userId {
		t.Error("error: user id is not correct")
	}

}
