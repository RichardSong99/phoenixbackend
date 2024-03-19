package quiz

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type QuizService struct {
	collection *mongo.Collection
}

func NewQuizService(ctx context.Context, client *mongo.Client) (*QuizService, error) {
	collection := client.Database("test").Collection("quizzes")

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

	return &QuizService{collection: collection}, nil
}

func (qs *QuizService) InitializeQuiz(ctx context.Context, questionIDs []primitive.ObjectID, userID primitive.ObjectID, quizType *string, quizName *string) (primitive.ObjectID, error) {
	// Create a new quiz
	quiz := &Quiz{
		UserID:           userID,
		QuestionIDList:   questionIDs,
		AttemptTime:      time.Now(),
		EngagementIDList: nil,
	}

	if quizType != nil {
		quiz.Type = *quizType
	}

	if quizName != nil {
		quiz.Name = *quizName
	}

	// Define the filter for the upsert operation
	filter := bson.M{"userID": quiz.UserID, "name": quiz.Name}

	// Define the options for the upsert operation
	replaceOptions := options.Replace().SetUpsert(true)

	// Perform the upsert operation
	replaceResult, err := qs.collection.ReplaceOne(ctx, filter, quiz, replaceOptions)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("error upserting quiz: %w", err)
	}

	// If a new document was inserted, return its ID
	if replaceResult.UpsertedID != nil {
		return replaceResult.UpsertedID.(primitive.ObjectID), nil
	}

	// If an existing document was replaced, return its ID
	return quiz.ID, nil
}

func (qs *QuizService) GetQuiz(ctx context.Context, quizID primitive.ObjectID) (*Quiz, error) {
	// Create a filter to find the quiz
	filter := bson.M{"_id": quizID}

	// Find the quiz
	var quiz Quiz
	err := qs.collection.FindOne(ctx, filter).Decode(&quiz)
	if err != nil {
		return nil, fmt.Errorf("error getting quiz: %w", err)
	}

	return &quiz, nil
}

func (qs *QuizService) GetQuizByName(ctx context.Context, name string, userID primitive.ObjectID) (*Quiz, error) {
	// Create a filter to find the quiz: by name and user ID
	filter := bson.M{"name": name, "user_id": userID}

	// Find the quiz
	var quiz Quiz
	err := qs.collection.FindOne(ctx, filter).Decode(&quiz)

	// if no quiz is found, return ErrNoDocuments
	if err == mongo.ErrNoDocuments {
		return nil, mongo.ErrNoDocuments
	}

	if err != nil {
		return nil, fmt.Errorf("error getting quiz by name: %w", err)
	}

	return &quiz, nil
}

func (qs *QuizService) GetQuizzesForUser(ctx context.Context, userID primitive.ObjectID) ([]*Quiz, error) {
	// Create a filter to find the quizzes for the user
	filter := bson.M{"user_id": userID}

	// Sort the quizzes by attempt time, with the earliest attempt time first
	opts := options.Find().SetSort(bson.D{{"attempt_time", 1}})

	// Find the quizzes
	cursor, err := qs.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting quizzes for user: %w", err)
	}

	// Iterate through the cursor and decode the quizzes
	var quizzes []*Quiz
	for cursor.Next(ctx) {
		var quiz Quiz
		err := cursor.Decode(&quiz)
		if err != nil {
			return nil, fmt.Errorf("error decoding quiz: %w", err)
		}
		quizzes = append(quizzes, &quiz)
	}

	return quizzes, nil
}

func (qs *QuizService) UpdateQuiz(ctx context.Context, quizID, engagementID primitive.ObjectID) (primitive.ObjectID, error) {
	// Create an update to add the engagement ID to the list of engagement IDs
	// do not add the same engagement ID to the list
	quiz, err := qs.GetQuiz(ctx, quizID)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("error getting quiz: %w", err)
	}

	// Check if the engagement ID already exists in the quiz
	for _, id := range quiz.EngagementIDList {
		if id == engagementID {
			// If engagement already exists, return the existing quiz ID without updating
			return quizID, nil
		}
	}

	// Define the filter to update the quiz
	filter := bson.M{"_id": quizID}

	// Define the update operation
	update := bson.M{"$push": bson.M{"engagement_id_list": engagementID}}

	// Update the quiz
	_, err = qs.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("error updating quiz: %w", err)
	}

	return quizID, nil
}
