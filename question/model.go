package question

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Question struct {
	ID                    *primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Prompt                *string             `bson:"prompt,omitempty" json:"Prompt,omitempty"`
	AnswerType            *string             `bson:"answer_type,omitempty" json:"AnswerType,omitempty"`
	AnswerChoices         *[]string           `bson:"answer_choices,omitempty" json:"AnswerChoices,omitempty"`
	CorrectAnswerMultiple *string             `bson:"correct_answer_multiple,omitempty" json:"CorrectAnswerMultiple,omitempty"`
	CorrectAnswerFree     *string             `bson:"correct_answer_free,omitempty" json:"CorrectAnswerFree,omitempty"`
	Text                  *string             `bson:"text,omitempty" json:"Text,omitempty"`
	Subject               *string             `bson:"subject,omitempty" json:"Subject,omitempty"`
	Topic                 *string             `bson:"topic,omitempty" json:"Topic,omitempty"`
	Difficulty            *string             `bson:"difficulty,omitempty" json:"Difficulty,omitempty"`
	AccessOption          *string             `bson:"access_option,omitempty" json:"AccessOption,omitempty"`
	Explanation           *string             `bson:"explanation,omitempty" json:"Explanation,omitempty"`
	Images                *[]Image            `bson:"images,omitempty" json:"Images,omitempty"`
	CreationDate          time.Time           `bson:"creation_date,omitempty" json:"CreationDate,omitempty"`
	LastEditedDate        time.Time           `bson:"last_edited_date,omitempty" json:"LastEditedDate,omitempty"`
}

type Image struct {
	Filename string `json:"Filename"`
	URL      string `json:"Url"`
}

type QuestionWithStatus struct {
	*Question
	Status *string `bson:"status" json:"Status"`
}
