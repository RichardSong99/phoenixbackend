package engagement

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// type Attempt struct {
// 	UserAnswer  *string       `bson:"user_answer,omitempty" json:"UserAnswer,omitempty"`
// 	Omitted     *bool         `bson:"omitted,omitempty" json:"Omitted,omitempty"`
// 	IsCorrect   *bool         `bson:"is_correct,omitempty" json:"IsCorrect,omitempty"`
// 	IsIncorrect *bool         `bson:"is_incorrect,omitempty" json:"IsIncorrect,omitempty"`
// 	AttemptTime time.Time     `bson:"attempt_time,omitempty" json:"AttemptTime,omitempty"`
// 	Duration    time.Duration `bson:"duration,omitempty" json:"Duration,omitempty"`
// 	Mode        *string       `bson:"mode,omitempty" json:"Mode,omitempty"`
// }

type Engagement struct {
	ID          *primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	QuestionID  *primitive.ObjectID `bson:"question_id,omitempty" json:"QuestionID,omitempty"`
	UserID      *primitive.ObjectID `bson:"user_id,omitempty" json:"UserID,omitempty"`
	Flagged     *bool               `bson:"flagged,omitempty" json:"Flagged,omitempty"`
	UserAnswer  *string             `bson:"user_answer,omitempty" json:"UserAnswer,omitempty"`
	Status      *string             `bson:"status,omitempty" json:"Status,omitempty"`
	AttemptTime time.Time           `bson:"attempt_time,omitempty" json:"AttemptTime,omitempty"`
	Duration    time.Duration       `bson:"duration,omitempty" json:"Duration,omitempty"`
	Mode        *string             `bson:"mode,omitempty" json:"Mode,omitempty"`
	Starred     *bool               `bson:"starred,omitempty" json:"Starred,omitempty"`
	Reviewed    *bool               `bson:"reviewed,omitempty" json:"Reviewed,omitempty"`
}
