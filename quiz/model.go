package quiz

import (
	"time"

	"example/goserver/engagement"
	"example/goserver/question"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Quiz struct {
	ID               primitive.ObjectID   `json:"id,omitempty" bson:"_id,omitempty"`
	Name             string               `json:"Name,omitempty" bson:"name,omitempty"`
	Type             string               `json:"Type,omitempty" bson:"type,omitempty"`
	UserID           primitive.ObjectID   `json:"UserID,omitempty" bson:"user_id,omitempty"`
	AttemptTime      time.Time            `json:"AttemptTime,omitempty" bson:"attempt_time,omitempty"`
	QuestionIDList   []primitive.ObjectID `json:"QuestionIDList,omitempty" bson:"question_id_list,omitempty"`
	EngagementIDList []primitive.ObjectID `json:"EngagementIDList,omitempty" bson:"engagement_id_list,omitempty"`
}

type QuestionEngagementCombo struct {
	Question   *question.Question     `json:"Question"`
	Engagement *engagement.Engagement `"json:Engagement`
}

type QuizResult struct {
	Quiz            *Quiz
	Questions       []QuestionEngagementCombo
	NumTotal        int
	NumAnswered     int
	NumCorrect      int
	NumIncorrect    int
	NumOmitted      int
	NumUnattempted  int
	PercentAnswered float64
	PercentCorrect  float64
}
