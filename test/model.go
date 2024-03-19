package test

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Test struct {
	ID          primitive.ObjectID    `json:"id,omitempty" bson:"_id,omitempty"`
	QuizIDList  *[]primitive.ObjectID `json:"QuizIDList,omitempty" bson:"quiz_id_list,omitempty"`
	UserID      *primitive.ObjectID   `json:"UserID,omitempty" bson:"user_id,omitempty"`
	Name        *string               `json:"Name,omitempty" bson:"name,omitempty"`
	AttemptTime time.Time             `json:"AttemptTime,omitempty" bson:"attempt_time,omitempty"`
	Completed   bool                  `json:"Completed,omitempty" bson:"completed,omitempty"`
}

type TestStats struct {
	StatList []SmallStats `json:"StatList"`
}

type SmallStats struct {
	Name    string `json:"Name"`
	Total   int    `json:"Total"`
	Correct int    `json:"Correct"`
}
