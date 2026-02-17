package errors

import (
	"errors"

	"go.mongodb.org/mongo-driver/mongo"
)

func IsNotFound(err error) bool {
	return errors.Is(err, mongo.ErrNoDocuments)
}
