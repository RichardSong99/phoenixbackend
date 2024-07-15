package user

import "go.mongodb.org/mongo-driver/bson/primitive"

var DefaultUserID primitive.ObjectID

func init() {
	var err error
	DefaultUserID, err = primitive.ObjectIDFromHex("000000000000000000000000")
	if err != nil {
		// This should never happen
		panic("Error creating ObjectID for unlogged user")
	}
}
