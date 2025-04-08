package utils

import (
	"testing"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func init() {
	err := godotenv.Load("../.env")
	if err != nil {
		panic("❌ Failed to load .env file")
	}
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
