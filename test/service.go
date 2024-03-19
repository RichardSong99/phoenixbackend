package test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TestService struct {
	collection *mongo.Collection
}

func NewTestService(ctx context.Context, client *mongo.Client) (*TestService, error) {
	collection := client.Database("test").Collection("tests")

	// Create the unique index
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "user_id", Value: 1}, // index in ascending order
			{Key: "name", Value: 1},    // index in ascending order
		},
		Options: options.Index().SetUnique(true),
	}
	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return nil, fmt.Errorf("could not create index: %w", err)
	}

	return &TestService{collection: collection}, nil
}

func (s *TestService) GetTestByName(c context.Context, name string, userID primitive.ObjectID) (*Test, error) {
	var test Test
	err := s.collection.FindOne(c, bson.M{"name": name, "user_id": userID}).Decode(&test)
	if err != nil {
		return nil, err
	}
	return &test, nil
}

func (s *TestService) GetTestByID(c context.Context, id primitive.ObjectID) (*Test, error) {
	var test Test
	err := s.collection.FindOne(c, bson.M{"_id": id}).Decode(&test)
	if err != nil {
		return nil, err
	}
	return &test, nil
}

func (s *TestService) CreateTest(c context.Context, quizIDList []primitive.ObjectID, name string, userID primitive.ObjectID) (primitive.ObjectID, error) {
	test := &Test{
		UserID:      &userID,
		QuizIDList:  &quizIDList,
		Name:        &name,
		AttemptTime: time.Now(),
		Completed:   false,
	}

	insertResult, err := s.collection.InsertOne(c, test)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return insertResult.InsertedID.(primitive.ObjectID), nil
}

func (s *TestService) GetTestsForUser(c context.Context, userID primitive.ObjectID) ([]Test, error) {
	cursor, err := s.collection.Find(c, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}

	var tests []Test
	if err = cursor.All(c, &tests); err != nil {
		return nil, err
	}

	return tests, nil
}

func (s *TestService) UpdateTest(c *gin.Context, id primitive.ObjectID, completed bool) error {
	result, err := s.collection.UpdateOne(
		c,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"Completed": completed}},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return fmt.Errorf("error updating test: %v", err)
	}

	if result.MatchedCount == 0 {
		return errors.New("no test found with the given ID")
	}

	return nil
}
