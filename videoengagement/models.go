package videoengagement

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// VideoEngagement represents a user's engagement with a video.
type VideoEngagement struct {
	ID      *primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	VideoID *primitive.ObjectID `bson:"video_id,omitempty" json:"VideoID,omitempty"`
	UserID  *primitive.ObjectID `bson:"user_id,omitempty" json:"UserID,omitempty"`
	Watched bool                `bson:"watched,omitempty" json:"Watched,omitempty"`
	Flagged bool                `bson:"flagged,omitempty" json:"Flagged,omitempty"`
}
