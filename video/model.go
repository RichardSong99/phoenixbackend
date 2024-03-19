package video

import "go.mongodb.org/mongo-driver/bson/primitive"

type Video struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	VideoID     string             `json:"VideoID,omitempty" bson:"video_id,omitempty"`
	Title       string             `json:"Title,omitempty" bson:"title,omitempty"`
	Description string             `json:"Description,omitempty" bson:"description,omitempty"`
	Thumbnail   string             `json:"Thumbnail,omitempty" bson:"thumbnail,omitempty"`
}

type VideoWithStatus struct {
	*Video
	Watched bool `json:"Watched,omitempty" bson:"watched,omitempty"`
	Flagged bool `json:"Flagged,omitempty" bson:"flagged,omitempty"`
}
