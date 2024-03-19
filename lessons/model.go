package lessons

import "go.mongodb.org/mongo-driver/bson/primitive"

type Lesson struct {
	ID   *primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name *string             `bson:"name,omitempty" json:"Name,omitempty"`
	Type *string             `bson:"type,omitempty" json:"Type,omitempty"`
	URL  *string             `bson:"url,omitempty" json:"URL,omitempty"`
}

type Course struct {
	ID        *primitive.ObjectID   `bson:"_id,omitempty" json:"id,omitempty"`
	Name      *string               `bson:"name,omitempty" json:"Name,omitempty"`
	LessonIDs []*primitive.ObjectID `bson:"lesson_ids,omitempty" json:"LessonIDs,omitempty"`
}
