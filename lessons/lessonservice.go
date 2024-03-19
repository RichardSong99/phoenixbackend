package lessons

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type LessonService struct {
	collection *mongo.Collection
}

func NewLessonService(client *mongo.Client) *LessonService {
	collection := client.Database("test").Collection("lessons")
	return &LessonService{collection: collection}
}

func (l *LessonService) AddLesson(ctx context.Context, lesson *Lesson) (*Lesson, error) {
	_, err := l.collection.InsertOne(ctx, lesson)
	if err != nil {
		return nil, err
	}
	return lesson, nil
}

func (l *LessonService) GetLessonByID(ctx context.Context, id string) (*Lesson, error) {
	var lesson Lesson

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	err = l.collection.FindOne(ctx, bson.M{"_id": oid}).Decode(&lesson)
	if err != nil {
		return nil, err
	}
	return &lesson, nil
}

func (l *LessonService) UpdateLesson(ctx context.Context, id string, lesson *Lesson) (*Lesson, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	_, err = l.collection.ReplaceOne(ctx, bson.M{"_id": oid}, lesson)
	if err != nil {
		return nil, err
	}
	return lesson, nil
}
