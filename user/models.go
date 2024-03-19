package user

import "go.mongodb.org/mongo-driver/bson/primitive"

// User represents a user in the system.
type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Email        string             `bson:"email" json:"Email"`
	PasswordHash string             `bson:"password_hash" json:"PasswordHash"`
	FirstName    string             `bson:"first_name" json:"FirstName"`
	LastName     string             `bson:"last_name" json:"LastName"`
	PhoneNumber  string             `bson:"phone_number" json:"PhoneNumber"`
	Tier         string             `bson:"tier" json:"Tier"`
}

