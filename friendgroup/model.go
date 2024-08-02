package friendgroup

import "go.mongodb.org/mongo-driver/bson/primitive"

// FriendGroup represents a friend group in the system.
type FriendGroup struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string             `bson:"name" json:"Name"`
	Description string             `bson:"description" json:"Description"`
	OwnerID     primitive.ObjectID `bson:"owner_id" json:"OwnerID"`

	// Members is a list of user IDs that are members of the friend group.
	Members []primitive.ObjectID `bson:"members" json:"Members"`
}
