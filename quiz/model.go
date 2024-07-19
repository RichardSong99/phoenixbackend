package quiz

import (
	"time"

	"example/goserver/engagement"
	"example/goserver/question"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Quiz struct {
	ID                         primitive.ObjectID          `json:"id,omitempty" bson:"_id,omitempty"`
	Name                       string                      `json:"name,omitempty" bson:"name,omitempty"`
	Type                       string                      `json:"type,omitempty" bson:"type,omitempty"`
	UserID                     primitive.ObjectID          `json:"userID,omitempty" bson:"user_id,omitempty"`
	AttemptTime                time.Time                   `json:"attempt_time,omitempty" bson:"attempt_time,omitempty"`
	QuestionEngagementIDCombos []QuestionEngagementIDCombo `json:"question_engagement_combos,omitempty" bson:"question_engagement_combos,omitempty"`
}

type QuestionEngagementCombo struct {
	Question   *question.Question     `json:"Question"`
	Engagement *engagement.Engagement `"json:Engagement`
}

type QuestionEngagementIDCombo struct {
	QuestionID   *primitive.ObjectID `json:"question_id" bson:"question_id"`
	EngagementID *primitive.ObjectID `json:"engagement_id" bson:"engagement_id"`
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
