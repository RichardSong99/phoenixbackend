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
		UserID:      userID,
		AttemptTime: time.Now(),
	}

	if quizType != nil {
		quiz.Type = *quizType
	}

	if quizName != nil {
		quiz.Name = *quizName
	}

	// Create the question-engagement ID combos
	// Initialize the slice if it doesn't exist
	if quiz.QuestionEngagementIDCombos == nil {
		quiz.QuestionEngagementIDCombos = []QuestionEngagementIDCombo{}
	}

	// Add new question IDs to the quiz
	for _, questionID := range questionIDs {
		found := false
		for _, qeid := range quiz.QuestionEngagementIDCombos {
			if qeid.QuestionID == &questionID {
				found = true
				break
			}
		}
		if !found {
			quiz.QuestionEngagementIDCombos = append(quiz.QuestionEngagementIDCombos, QuestionEngagementIDCombo{
				QuestionID: &questionID,
			})
		}
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

func (qs *QuizService) GetQuizzesForUser(ctx context.Context, userID primitive.ObjectID, quizType *string) ([]*Quiz, error) {
	// Create a filter to find the quizzes for the user
	filter := bson.M{"user_id": userID}

	// if quizType is not empty, add it to the filter
	if quizType != nil && *quizType != "" {
		filter["type"] = *quizType
	}

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

func (qs *QuizService) UpdateQuiz(ctx context.Context, quizID, questionID, engagementID *primitive.ObjectID) (primitive.ObjectID, error) {
	// Fetch the quiz
	quiz, err := qs.GetQuiz(ctx, *quizID)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("error getting quiz: %w", err)
	}

	// Check if the question ID is in the quiz
	questionFound := false
	for i, qeid := range quiz.QuestionEngagementIDCombos {
		if qeid.QuestionID == questionID {
			questionFound = true

			// Check if the engagement ID is already associated with this question
			if qeid.EngagementID == nil || qeid.EngagementID != engagementID {
				// Update the engagement ID for the existing question
				quiz.QuestionEngagementIDCombos[i].EngagementID = engagementID
			}
			break
		}
	}

	if !questionFound {
		// Add a new entry if the question ID was not found
		quiz.QuestionEngagementIDCombos = append(quiz.QuestionEngagementIDCombos, QuestionEngagementIDCombo{
			QuestionID:   questionID,
			EngagementID: engagementID,
		})
	}

	// Define the filter to update the quiz
	filter := bson.M{"_id": quizID}

	// Define the update operation
	update := bson.M{
		"$set": bson.M{
			"question_engagement_combos": quiz.QuestionEngagementIDCombos,
		},
	}

	// Update the quiz in the database
	_, err = qs.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("error updating quiz: %w", err)
	}

	return *quizID, nil
}

func (qs *QuizService) UpdateQuizWithCombos(ctx context.Context, quizID primitive.ObjectID, qeidCombos []QuestionEngagementIDCombo) (primitive.ObjectID, error) {
	// Define the filter to update the quiz
	var quiz Quiz

	if err := qs.collection.FindOne(ctx, bson.M{"_id": quizID}).Decode(&quiz); err != nil {
		return primitive.NilObjectID, fmt.Errorf("error finding quiz: %w", err)
	}

	// Map of question IDs to engagement IDs for fast lookup
	// comboMap := make(map[primitive.ObjectID]*primitive.ObjectID)
	// for _, combo := range quiz.QuestionEngagementIDCombos {
	// 	comboMap[*combo.QuestionID] = combo.EngagementID
	// }

	// // update the map with the new combos
	// for _, combo := range qeidCombos {
	// 	comboMap[*combo.QuestionID] = combo.EngagementID
	// }

	// // rebuild the question-engagement ID combos
	// // rebuild the question-engagement ID combos
	// var newCombos []QuestionEngagementIDCombo
	// for questionID, engagementID := range comboMap {
	// 	qIDCopy := questionID // Avoid referencing the same memory
	// 	var eIDCopy *primitive.ObjectID
	// 	if engagementID != nil {
	// 		eIDCopyCopy := *engagementID // Avoid referencing the same memory
	// 		eIDCopy = &eIDCopyCopy
	// 	}
	// 	newCombos = append(newCombos, QuestionEngagementIDCombo{
	// 		QuestionID:   &qIDCopy,
	// 		EngagementID: eIDCopy,
	// 	})
	// }

	// create a map to track existing question IDs
	existingQuestions := make(map[primitive.ObjectID]bool)
	for _, combo := range quiz.QuestionEngagementIDCombos {
		existingQuestions[*combo.QuestionID] = true
	}

	// append the new combos to the end of the list
	for _, combo := range qeidCombos {
		if _, ok := existingQuestions[*combo.QuestionID]; !ok {
			quiz.QuestionEngagementIDCombos = append(quiz.QuestionEngagementIDCombos, combo)
		} else {
			for i, qeid := range quiz.QuestionEngagementIDCombos {
				if qeid.QuestionID == combo.QuestionID {
					quiz.QuestionEngagementIDCombos[i].EngagementID = combo.EngagementID
				}
			}

		}
	}

	// Define the filter to update the quiz
	filter := bson.M{"_id": quizID}

	// Define the update operation
	update := bson.M{
		"$set": bson.M{
			"question_engagement_id_combos": quiz.QuestionEngagementIDCombos,
		},
	}

	// Update the quiz in the database
	_, err := qs.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("error updating quiz: %w", err)
	}

	return quizID, nil
}
