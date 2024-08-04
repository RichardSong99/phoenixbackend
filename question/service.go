package question

import (
	"context"
	"errors"

	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type QuestionService struct {
	collection *mongo.Collection
}

// Modify this function to remove the engagementService parameter
func NewQuestionService(client *mongo.Client) *QuestionService { // Modify this line
	collection := client.Database("test").Collection("questions")
	return &QuestionService{
		collection: collection,
		// Remove the engagementService field
	}
}

// ... (other methods here) ...
func (s *QuestionService) CreateQuestion(ctx context.Context, question *Question) (*mongo.InsertOneResult, error) {
	return s.collection.InsertOne(ctx, question)
}

func (s *QuestionService) GetQuestion(ctx context.Context, id primitive.ObjectID) (*Question, error) {
	var question Question
	err := s.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&question)
	if err != nil {
		return nil, err
	}
	return &question, nil
}

// GetAllQuestions retrieves all questions from the database
func (s *QuestionService) GetAllQuestions(c *gin.Context) ([]Question, error) {
	// Create a context for the operation
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find all documents in the collection
	cursor, err := s.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Decode the documents into a slice of Questions
	var questions []Question
	if err := cursor.All(ctx, &questions); err != nil {
		return nil, err
	}

	return questions, nil
}

func (s *QuestionService) GetQuestionByID(ctx context.Context, id primitive.ObjectID) (*Question, error) {
	var question Question
	err := s.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&question)
	if err != nil {
		return nil, err
	}
	return &question, nil
}

func (s *QuestionService) GetQuestionsByID(ctx context.Context, ids []primitive.ObjectID) ([]Question, error) {
	cursor, err := s.collection.Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		return nil, err
	}

	var questions []Question
	if err = cursor.All(ctx, &questions); err != nil {
		return nil, err
	}

	return questions, nil
}

// GetQuestions retrieves questions from the database
// based on the provided difficulty, topic, and limit
func (s *QuestionService) GetQuestions(ctx context.Context, difficulties string, topics string, answerStatus string, answerType string, skip, pageSize int64, userTier string, userID *primitive.ObjectID, subject string, sortOption string, sortDirection string) ([]bson.M, int64, error) {

	filter := s.createFilter(difficulties, topics, answerType, subject)

	// Create the initial pipeline with the match stage
	pipeline := []bson.M{
		{"$match": filter},
	}

	// add user engagement filter
	if userID != nil {
		pipeline = s.addUserEngagementFilter(pipeline, userID)
	}

	pipeline = s.addFirstAttemptTimeToPipeline(pipeline)

	fmt.Println("Pipeline after add first attempt time: ", pipeline)

	// add difficulty levels
	pipeline = s.addDifficultyLevelsToPipeline(pipeline)

	pipeline = s.addAnswerStatusToPipeline(pipeline)

	// fmt.Println("answerStatus: ", answerStatus)

	if answerStatus != "" {
		pipeline = s.addAnswerStatusFilterToPipeline(pipeline, answerStatus)
	}

	pipeline = s.addMatchAndSortStagesToPipeline(pipeline, sortOption, sortDirection)

	pipeline = append(pipeline, generateProjectStage())

	countPipeline := make([]bson.M, len(pipeline))
	copy(countPipeline, pipeline)

	totalQuestions, err := s.getTotalQuestions(ctx, countPipeline)
	if err != nil {
		return nil, 0, err
	}

	pipeline = s.addPagination(pipeline, skip, pageSize)

	// pipeline = s.addProjectionStage(pipeline)

	fmt.Println("Pipeline: ", pipeline)

	results, err := s.executePipelineGeneric(ctx, pipeline)
	if err != nil {
		return nil, 0, err
	}

	return results, totalQuestions, nil
}

func (s *QuestionService) createFilter(difficulties string, topics string, answerType string, subject string) bson.M {
	filter := bson.M{}
	if difficulties != "" {
		difficultySlice := strings.Split(strings.ToLower(difficulties), ",")
		filter["difficulty"] = bson.M{"$in": difficultySlice}
	}
	if topics != "" {
		topicSlice := strings.Split(topics, ",")
		filter["topic"] = bson.M{"$in": topicSlice}
	}
	if answerType != "" {
		answerTypeSlice := strings.Split(answerType, ",")
		filter["answer_type"] = bson.M{"$in": answerTypeSlice}
	}
	if subject != "" {
		filter["subject"] = subject
	}
	return filter
}

func (s *QuestionService) addUserEngagementFilter(pipeline []bson.M, userID *primitive.ObjectID) []bson.M {
	pipeline = append(pipeline,
		bson.M{
			"$lookup": bson.M{
				"from":         "engagements",
				"localField":   "_id",
				"foreignField": "question_id",
				"as":           "engagements",
			},
		},
		bson.M{
			"$addFields": bson.M{
				"engagements": bson.M{
					"$filter": bson.M{
						"input": "$engagements",
						"as":    "engagement",
						"cond": bson.M{
							"$eq": []interface{}{"$$engagement.user_id", userID},
						},
					},
				},
			},
		},
	)

	return pipeline
}

func (s *QuestionService) addDifficultyLevelsToPipeline(pipeline []bson.M) []bson.M {
	pipeline = append(pipeline, bson.M{
		"$addFields": bson.M{
			"difficultyLevel": bson.M{
				"$switch": bson.M{
					"branches": []bson.M{
						{
							"case": bson.M{"$eq": bson.A{"$difficulty", "easy"}},
							"then": 1,
						},
						{
							"case": bson.M{"$eq": bson.A{"$difficulty", "medium"}},
							"then": 2,
						},
						{
							"case": bson.M{"$eq": bson.A{"$difficulty", "hard"}},
							"then": 3,
						},
						{
							"case": bson.M{"$eq": bson.A{"$difficulty", "extreme"}},
							"then": 4,
						},
					},
					"default": 0,
				},
			},
		},
	})

	return pipeline
}

func (s *QuestionService) addFirstAttemptTimeToPipeline(pipeline []bson.M) []bson.M {
	pipeline = append(pipeline, bson.M{
		"$addFields": bson.M{
			"first_attempt_time": bson.M{
				"$cond": bson.M{
					"if":   bson.M{"$gt": bson.A{bson.M{"$size": "$engagements"}, 0}},
					"then": bson.M{"$arrayElemAt": bson.A{"$engagements.attempt_time", 0}},
					"else": nil,
				},
			},
		},
	})

	return pipeline
}

func (s *QuestionService) addAnswerStatusToPipeline(pipeline []bson.M) []bson.M {
	pipeline = append(pipeline, bson.M{
		"$project": bson.M{
			"question": "$$ROOT",
			"status": bson.M{
				"$cond": bson.M{
					"if": bson.M{
						"$eq": bson.A{bson.M{"$size": "$engagements"}, 0}, // If engagements is empty
					}, "then": "unattempted",
					"else": bson.M{
						"$arrayElemAt": bson.A{"$engagements.status", 0},
					},
				},
			},
		},
	})

	return pipeline
}

func (s *QuestionService) addAnswerStatusFilterToPipeline(pipeline []bson.M, answerStatus string) []bson.M {
	selectedAnswerStatusArray := strings.Split(answerStatus, ",")
	filter := bson.M{
		"status": bson.M{
			"$in": selectedAnswerStatusArray,
		},
	}

	pipeline = append(pipeline, bson.M{"$match": filter})

	return pipeline
}

func (s *QuestionService) addMatchAndSortStagesToPipeline(pipeline []bson.M, sortOption string, sortDirection string) []bson.M {
	// pipeline = append(pipeline, bson.M{"$match": filter})

	// Add sort stage to the pipeline
	sortStage := bson.M{}

	sortText := ""

	if sortOption == "attemptTime" {
		sortText = "question.first_attempt_time"
	} else if sortOption == "createdTime" {
		sortText = "question.creation_date"
	} else if sortOption == "lastEditedTime" {
		sortText = "question.last_edited_date"
	}

	sortDirectionInt, err := strconv.Atoi(sortDirection)

	if err != nil {
		sortDirectionInt = 1
	}

	if sortText != "" {
		sortStage[sortText] = sortDirectionInt
	}

	if len(sortStage) > 0 {
		pipeline = append(pipeline, bson.M{"$sort": sortStage})
	}

	return pipeline
}

func (s *QuestionService) getTotalQuestions(ctx context.Context, countPipeline []bson.M) (int64, error) {
	// Add a count stage to the count pipeline
	countPipeline = append(countPipeline, bson.M{"$count": "total"})

	// Execute the count pipeline
	countCursor, err := s.collection.Aggregate(ctx, countPipeline)
	if err != nil {
		return 0, err
	}
	defer countCursor.Close(ctx)

	var counts []struct {
		Total int64 `bson:"total"`
	}
	if err = countCursor.All(ctx, &counts); err != nil {
		return 0, err
	}

	totalQuestions := int64(0)
	if len(counts) > 0 {
		totalQuestions = counts[0].Total
	}

	return totalQuestions, nil
}

func (s *QuestionService) addPagination(pipeline []bson.M, skip, pageSize int64) []bson.M {
	// Add skip and limit stages to the pipeline for pagination
	if skip > 0 {
		pipeline = append(pipeline, bson.M{"$skip": skip})
	}
	if pageSize > 0 {
		pipeline = append(pipeline, bson.M{"$limit": pageSize})
	}

	return pipeline
}

func (s *QuestionService) addProjectionStage(pipeline []bson.M) []bson.M {
	// Add projection stage to the pipeline
	pipeline = append(pipeline, bson.M{
		"$project": bson.M{
			"question": "$$ROOT",
			"status":   "$status",
		},
	})

	return pipeline
}

func (s *QuestionService) executePipelineGeneric(ctx context.Context, pipeline []bson.M) ([]bson.M, error) {
	cursor, err := s.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

func generateProjectStage() bson.M {

	return bson.M{
		"$project": bson.M{
			"Question": bson.M{
				"id":                    "$question._id",
				"Prompt":                "$question.prompt",
				"AnswerType":            "$question.answer_type",
				"AnswerChoices":         "$question.answer_choices",
				"CorrectAnswerMultiple": "$question.correct_answer_multiple",
				"CorrectAnswerFree":     "$question.correct_answer_free",
				"Text":                  "$question.text",
				"Subject":               "$question.subject",
				"Topic":                 "$question.topic",
				"Difficulty":            "$question.difficulty",
				"AccessOption":          "$question.access_option",
				"Explanation":           "$question.explanation",
				"Images":                "$question.images",
				"CreationDate":          "$question.creation_date",
				"LastEditedDate":        "$question.last_edited_date",
				"DifficultyLevel":       "$question.difficultyLevel",
			},
			"Engagement": bson.M{
				"$cond": bson.M{
					"if": bson.M{"$gt": []interface{}{bson.M{"$size": "$question.engagements"}, 0}},
					"then": bson.M{
						"id":          bson.M{"$arrayElemAt": []interface{}{"$question.engagements._id", 0}},
						"QuestionID":  bson.M{"$arrayElemAt": []interface{}{"$question.engagements.question_id", 0}},
						"UserID":      bson.M{"$arrayElemAt": []interface{}{"$question.engagements.user_id", 0}},
						"Flagged":     bson.M{"$arrayElemAt": []interface{}{"$question.engagements.flagged", 0}},
						"UserAnswer":  bson.M{"$arrayElemAt": []interface{}{"$question.engagements.user_answer", 0}},
						"Status":      bson.M{"$arrayElemAt": []interface{}{"$question.engagements.status", 0}},
						"AttemptTime": bson.M{"$arrayElemAt": []interface{}{"$question.engagements.attempt_time", 0}},
						"Duration":    bson.M{"$arrayElemAt": []interface{}{"$question.engagements.duration", 0}},
						"Mode":        bson.M{"$arrayElemAt": []interface{}{"$question.engagements.mode", 0}},
					},
					"else": nil,
				},
			},
			"status": "$status",
		},
	}
}

// GetMaskedQuestions retrieves questions with specific fields from the database
func (s *QuestionService) GetMaskedQuestions(ctx context.Context, jsonFields string) ([]bson.M, error) {
	structType := reflect.TypeOf(Question{})
	projection := bson.M{}

	for _, jsonField := range strings.Split(jsonFields, ",") {
		jsonField = strings.TrimSpace(jsonField) // Trim any extra whitespace
		if bsonField, ok := jsonToBsonFieldName(structType, jsonField); ok {
			projection[bsonField] = 1
		}
	}

	// Debug: Print the projection to check if it's correctly constructed
	fmt.Printf("Projection: %+v\n", projection)

	// MongoDB query using the projection
	cursor, err := s.collection.Find(ctx, bson.M{}, options.Find().SetProjection(projection))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var questions []bson.M
	if err := cursor.All(ctx, &questions); err != nil {
		return nil, err
	}

	return questions, nil
}

// UpdateQuestion updates a question in the database
func (s *QuestionService) UpdateQuestion(ctx context.Context, id primitive.ObjectID, update bson.M) (*mongo.UpdateResult, error) {
	return s.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
}

func (s *QuestionService) UpdateAllQuestions(ctx context.Context, update bson.M) (*mongo.UpdateResult, error) {
	// Check if update is empty
	if len(update) == 0 {
		return nil, errors.New("update cannot be empty")
	}

	// Check if the value of the field is an empty string
	for _, value := range update {
		if str, ok := value.(string); ok && str == "" {
			return nil, errors.New("update value cannot be an empty string")
		}
	}

	filter := bson.M{} // This is an empty filter which will match all documents in the collection.
	return s.collection.UpdateMany(ctx, filter, bson.M{"$set": update})
}

// DeleteQuestion deletes a question from the database

func (s *QuestionService) DeleteQuestion(ctx context.Context, id primitive.ObjectID) (*mongo.DeleteResult, error) {
	return s.collection.DeleteOne(ctx, bson.M{"_id": id})
}

// jsonToBsonFieldName finds the BSON field name for a given JSON field name.
func jsonToBsonFieldName(structType reflect.Type, jsonName string) (string, bool) {
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		jsonTag := field.Tag.Get("json")
		bsonTag := field.Tag.Get("bson")

		jsonFieldName := strings.Split(jsonTag, ",")[0]
		bsonFieldName := strings.Split(bsonTag, ",")[0]

		fmt.Printf("Field: %s, JSON: %s, BSON: %s\n", field.Name, jsonFieldName, bsonFieldName)

		if jsonFieldName == jsonName {
			return bsonFieldName, true
		}
	}
	return "", false
}
