package question

import (
	"context"
	"errors"
	"example/goserver/dataaggregation"
	"example/goserver/user"
	"fmt"
	"math/rand"
	"reflect"
	"sort"
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

func (s *QuestionService) GetQuestionsByID(ctx context.Context, questionids []primitive.ObjectID, userID *primitive.ObjectID) ([]*QuestionWithStatus, error) {
	// Create the initial pipeline with the match stage
	pipeline := []bson.M{
		{"$match": bson.M{"_id": bson.M{"$in": questionids}}},
	}

	// Add the other stages to the pipeline
	pipeline = append(pipeline, s.createInitialPipeline(userID)...)

	// Add the facet stage to the pipeline
	answerStatus := "unattempted,incorrect,omitted,correct,flagged"

	pipeline = s.addFacetStageToPipeline(pipeline, userID, answerStatus)
	pipeline = s.addProjectionStage(pipeline)

	results, err := s.executePipeline(ctx, pipeline)
	if err != nil {
		return nil, err
	}

	// Create a map with the ObjectID as the key and the corresponding index as the value
	idIndexMap := make(map[primitive.ObjectID]int)
	for i, id := range questionids {
		idIndexMap[id] = i
	}

	// Sort the results based on the order of questionids
	sort.Slice(results, func(i, j int) bool {
		return idIndexMap[*results[i].ID] < idIndexMap[*results[j].ID]
	})

	return results, nil
}

func (s *QuestionService) GetDifficultyStatistics(ctx context.Context, userID *primitive.ObjectID) (interface{}, error) {
	pipeline := s.createInitialPipeline(userID)

	answerStatus := "unattempted,incorrect,omitted,correct,flagged"

	if userID != nil {
		pipeline = s.addFacetStageToPipeline(pipeline, userID, answerStatus)
	}

	difficultyPipeline := s.createDifficultyPipeline(pipeline)
	difficultyResults, err := s.executeStatsPipeline(ctx, difficultyPipeline)
	if err != nil {
		return nil, err
	}

	return difficultyResults, nil
}

func (s *QuestionService) GetStatusStatistics(ctx context.Context, userID *primitive.ObjectID) (interface{}, error) {
	pipeline := s.createInitialPipeline(userID)

	answerStatus := "unattempted,incorrect,omitted,correct,flagged"

	if userID != nil {
		pipeline = s.addFacetStageToPipeline(pipeline, userID, answerStatus)
	}

	statusPipeline := s.createStatusPipeline(pipeline)
	statusResults, err := s.executeStatsPipeline(ctx, statusPipeline)
	if err != nil {
		return nil, err
	}

	return statusResults, nil
}

func (s *QuestionService) GetCombinedStatistics(ctx context.Context, userID *primitive.ObjectID) ([]dataaggregation.TopicStat, error) {
	pipeline := s.createInitialPipeline(userID)

	answerStatus := "unattempted,incorrect,omitted,correct,flagged"

	if userID != nil {
		pipeline = s.addFacetStageToPipeline(pipeline, userID, answerStatus)
	}

	combinedPipeline := s.createCombinedPipeline(pipeline)
	combinedResults, err := s.executeCombinedStatsPipeline(ctx, combinedPipeline)
	if err != nil {
		return nil, err
	}

	return combinedResults, nil
}

func (s *QuestionService) GetCombinedCubeStatistics(ctx context.Context, userID *primitive.ObjectID) ([]dataaggregation.TopicAggregation, error) {
	pipeline := s.createInitialPipeline(userID)

	answerStatus := "unattempted,incorrect,omitted,correct,flagged"

	if userID != nil {
		pipeline = s.addFacetStageToPipeline(pipeline, userID, answerStatus)
	}

	combinedPipeline := s.createCombinedCubePipeline(pipeline)
	combinedResults, err := s.executeCombinedCubeStatsPipeline(ctx, combinedPipeline)
	if err != nil {
		return nil, err
	}

	return combinedResults, nil
}

func (s *QuestionService) GetTimeStatistics(ctx context.Context, userID *primitive.ObjectID) (interface{}, error) {
	pipeline := s.createInitialPipeline(userID)

	answerStatus := "unattempted,incorrect,omitted,correct,flagged"

	if userID != nil {
		pipeline = s.addFacetStageToPipeline(pipeline, userID, answerStatus)
	}

	timePipeline := s.createTimePipeline(pipeline)
	timeResults, err := s.executeStatsPipeline(ctx, timePipeline)
	if err != nil {
		return nil, err
	}

	return timeResults, nil
}

func (s *QuestionService) createDifficultyPipeline(pipeline []bson.M) []bson.M {
	difficultyPipeline := []bson.M{
		{
			"$group": bson.M{
				"_id": bson.M{
					"topic":      "$topic",
					"difficulty": "$difficulty",
				},
				"count": bson.M{"$sum": 1},
			},
		},
		{
			"$group": bson.M{
				"_id": "$_id.topic",
				"difficulties": bson.M{
					"$push": bson.M{
						"k": "$_id.difficulty",
						"v": "$count",
					},
				},
				"total": bson.M{"$sum": "$count"}, // Add this line

			},
		},
		{
			"$addFields": bson.M{
				"difficulties": bson.M{
					"$arrayToObject": "$difficulties",
				},
			},
		},
		{
			"$replaceRoot": bson.M{
				"newRoot": bson.M{
					"$mergeObjects": []interface{}{
						bson.M{"topic": "$_id"},
						"$difficulties",
						bson.M{"total": "$total"}, // Add this line

					},
				},
			},
		},
	}

	difficultyPipeline = append(pipeline, difficultyPipeline...)
	return difficultyPipeline
}

func (s *QuestionService) createStatusPipeline(pipeline []bson.M) []bson.M {
	statusPipeline := []bson.M{
		{
			"$group": bson.M{
				"_id": bson.M{
					"topic": "$topic",
				},
				"unattempted": bson.M{"$sum": bson.M{"$cond": bson.M{"if": bson.M{"$eq": []interface{}{"$status", "unattempted"}}, "then": 1, "else": 0}}},
				"incorrect":   bson.M{"$sum": bson.M{"$cond": bson.M{"if": bson.M{"$eq": []interface{}{"$status", "incorrect"}}, "then": 1, "else": 0}}},
				"omitted":     bson.M{"$sum": bson.M{"$cond": bson.M{"if": bson.M{"$eq": []interface{}{"$status", "omitted"}}, "then": 1, "else": 0}}},
				"correct":     bson.M{"$sum": bson.M{"$cond": bson.M{"if": bson.M{"$eq": []interface{}{"$status", "correct"}}, "then": 1, "else": 0}}},
				"flagged":     bson.M{"$sum": bson.M{"$cond": bson.M{"if": bson.M{"$eq": []interface{}{"$status", "flagged"}}, "then": 1, "else": 0}}},
				"total":       bson.M{"$sum": 1},
			},
		},
		{
			"$project": bson.M{
				"topic":       "$_id.topic",
				"unattempted": 1,
				"incorrect":   1,
				"omitted":     1,
				"correct":     1,
				"flagged":     1,
				"total":       1,

				"_id": 0,
			},
		},
	}

	statusPipeline = append(pipeline, statusPipeline...)
	return statusPipeline
}

func (s *QuestionService) createCombinedPipeline(pipeline []bson.M) []bson.M {
	combinedPipeline := []bson.M{
		{
			"$group": bson.M{
				"_id": bson.M{
					"topic":      "$topic",
					"difficulty": "$difficulty",
				},
				"count":       bson.M{"$sum": 1},
				"unattempted": bson.M{"$sum": bson.M{"$cond": bson.M{"if": bson.M{"$eq": []interface{}{"$status", "unattempted"}}, "then": 1, "else": 0}}},
				"incorrect":   bson.M{"$sum": bson.M{"$cond": bson.M{"if": bson.M{"$eq": []interface{}{"$status", "incorrect"}}, "then": 1, "else": 0}}},
				"omitted":     bson.M{"$sum": bson.M{"$cond": bson.M{"if": bson.M{"$eq": []interface{}{"$status", "omitted"}}, "then": 1, "else": 0}}},
				"correct":     bson.M{"$sum": bson.M{"$cond": bson.M{"if": bson.M{"$eq": []interface{}{"$status", "correct"}}, "then": 1, "else": 0}}},
				"flagged":     bson.M{"$sum": bson.M{"$cond": bson.M{"if": bson.M{"$eq": []interface{}{"$status", "flagged"}}, "then": 1, "else": 0}}},
			},
		},
		{
			"$group": bson.M{
				"_id": "$_id.topic",
				"difficulties": bson.M{
					"$push": bson.M{
						"k": "$_id.difficulty",
						"v": bson.M{
							"total":       "$count",
							"unattempted": "$unattempted",
							"incorrect":   "$incorrect",
							"omitted":     "$omitted",
							"correct":     "$correct",
							"flagged":     "$flagged",
						},
					},
				},
				"totalCount":       bson.M{"$sum": "$count"},
				"totalUnattempted": bson.M{"$sum": "$unattempted"},
				"totalIncorrect":   bson.M{"$sum": "$incorrect"},
				"totalOmitted":     bson.M{"$sum": "$omitted"},
				"totalCorrect":     bson.M{"$sum": "$correct"},
				"totalFlagged":     bson.M{"$sum": "$flagged"},
			},
		},
		{
			"$addFields": bson.M{
				"difficulties": bson.M{
					"$arrayToObject": "$difficulties",
				},
				"total": bson.M{
					"total": bson.M{
						"total":       "$totalCount",
						"unattempted": "$totalUnattempted",
						"incorrect":   "$totalIncorrect",
						"omitted":     "$totalOmitted",
						"correct":     "$totalCorrect",
						"flagged":     "$totalFlagged",
					},
				},
			},
		},
		{
			"$replaceRoot": bson.M{
				"newRoot": bson.M{
					"$mergeObjects": []interface{}{
						bson.M{"topic": "$_id"},
						"$difficulties",
						"$total",
					},
				},
			},
		},
	}

	combinedPipeline = append(pipeline, combinedPipeline...)
	return combinedPipeline
}

func (s *QuestionService) createCombinedCubePipeline(pipeline []bson.M) []bson.M {
	combinedPipeline := []bson.M{
		{
			"$group": bson.M{
				"_id": bson.M{
					"topic":      "$topic",
					"status":     "$status",
					"difficulty": "$difficulty",
				},
				"count": bson.M{"$sum": 1},
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"topic":  "$_id.topic",
					"status": "$_id.status",
				},
				"difficulties": bson.M{
					"$push": bson.M{
						"k": "$_id.difficulty",
						"v": "$count",
					},
				},
			},
		},
		{
			"$group": bson.M{
				"_id": "$_id.topic",
				"statuses": bson.M{
					"$push": bson.M{
						"k": "$_id.status",
						"v": bson.M{
							"difficulties": "$difficulties",
						},
					},
				},
			},
		},
		{
			"$addFields": bson.M{
				"statuses": bson.M{
					"$arrayToObject": "$statuses",
				},
			},
		},
		{
			"$addFields": bson.M{
				"statuses": bson.M{
					"$objectToArray": "$statuses",
				},
			},
		},
		{
			"$unwind": "$statuses",
		},
		{
			"$addFields": bson.M{
				"statuses.v.difficulties": bson.M{
					"$arrayToObject": "$statuses.v.difficulties",
				},
			},
		},
		{
			"$group": bson.M{
				"_id": "$_id",
				"statuses": bson.M{
					"$push": bson.M{
						"k": "$statuses.k",
						"v": bson.M{
							"difficulties": "$statuses.v.difficulties",
						},
					},
				},
			},
		},
		{
			"$addFields": bson.M{
				"statuses": bson.M{
					"$arrayToObject": "$statuses",
				},
			},
		},
		{
			"$replaceRoot": bson.M{
				"newRoot": bson.M{
					"$mergeObjects": []interface{}{
						bson.M{"topic": "$_id"},
						bson.M{"statuses": "$statuses"},
					},
				},
			},
		},
	}

	combinedPipeline = append(pipeline, combinedPipeline...)
	return combinedPipeline
}

func (s *QuestionService) createTimePipeline(pipeline []bson.M) []bson.M {
	timePipeline := []bson.M{
		{
			"$match": bson.M{
				"first_attempt_time": bson.M{
					"$ne": nil,
				},
			},
		},
		{
			"$addFields": bson.M{
				"date": bson.M{
					"$dateToString": bson.M{
						"format": "%Y-%m-%d",
						"date":   "$first_attempt_time",
					},
				},
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"date":       "$date",
					"topic":      "$topic",
					"difficulty": "$difficulty",
				},
				"count":     bson.M{"$sum": 1},
				"incorrect": bson.M{"$sum": bson.M{"$cond": bson.M{"if": bson.M{"$eq": []interface{}{"$status", "incorrect"}}, "then": 1, "else": 0}}},
				"omitted":   bson.M{"$sum": bson.M{"$cond": bson.M{"if": bson.M{"$eq": []interface{}{"$status", "omitted"}}, "then": 1, "else": 0}}},
				"correct":   bson.M{"$sum": bson.M{"$cond": bson.M{"if": bson.M{"$eq": []interface{}{"$status", "correct"}}, "then": 1, "else": 0}}},
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"date":  "$_id.date",
					"topic": "$_id.topic",
				},
				"difficulties": bson.M{
					"$push": bson.M{
						"difficulty": "$_id.difficulty",
						"results": bson.M{
							"total":     "$count",
							"incorrect": "$incorrect",
							"omitted":   "$omitted",
							"correct":   "$correct",
						},
					},
				},
			},
		},
		{
			"$group": bson.M{
				"_id": "$_id.date",
				"topics": bson.M{
					"$push": bson.M{
						"k": "$_id.topic",
						"v": "$difficulties",
					},
				},
			},
		},
		{
			"$addFields": bson.M{
				"topics": bson.M{
					"$arrayToObject": "$topics",
				},
			},
		},
		{
			"$replaceRoot": bson.M{
				"newRoot": bson.M{
					"$mergeObjects": []interface{}{
						bson.M{"date": "$_id"},
						"$topics",
					},
				},
			},
		},
	}

	timePipeline = append(pipeline, timePipeline...)
	return timePipeline
}

func (s *QuestionService) executeStatsPipeline(ctx context.Context, pipeline []bson.M) ([]bson.M, error) {
	cursor, err := s.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	// fmt.Printf("Results: %+v\n", results)

	return results, nil
}

func (s *QuestionService) executeCombinedStatsPipeline(ctx context.Context, pipeline []bson.M) ([]dataaggregation.TopicStat, error) {
	cursor, err := s.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	var topicStats []dataaggregation.TopicStat
	for _, result := range results {
		var topicStat dataaggregation.TopicStat
		bsonBytes, _ := bson.Marshal(result)
		bson.Unmarshal(bsonBytes, &topicStat)
		topicStats = append(topicStats, topicStat)
	}

	return topicStats, nil
}

func (s *QuestionService) executeCombinedCubeStatsPipeline(ctx context.Context, pipeline []bson.M) ([]dataaggregation.TopicAggregation, error) {
	cursor, err := s.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	var topicStats []dataaggregation.TopicAggregation
	for _, result := range results {
		var topicStat dataaggregation.TopicAggregation
		bsonBytes, _ := bson.Marshal(result)
		bson.Unmarshal(bsonBytes, &topicStat)
		topicStats = append(topicStats, topicStat)
	}

	return topicStats, nil
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

	// add difficulty levels
	pipeline = s.addDifficultyLevelsToPipeline(pipeline)

	// pipeline := s.createInitialPipeline(userID)

	// if userID != nil {
	// 	pipeline = s.addFacetStageToPipeline(pipeline, userID, answerStatus)
	// }

	pipeline = s.addAnswerStatusToPipeline(pipeline)

	fmt.Println("answerStatus: ", answerStatus)

	if answerStatus != "" {
		pipeline = s.addAnswerStatusFilterToPipeline(pipeline, answerStatus)
	}

	fmt.Println("Pipeline: ", pipeline)

	pipeline = s.addMatchAndSortStagesToPipeline(pipeline, sortOption, sortDirection)

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

func (s *QuestionService) createInitialPipeline(userID *primitive.ObjectID) []bson.M {
	return []bson.M{
		{
			"$lookup": bson.M{
				"from":         "engagements",
				"localField":   "_id",
				"foreignField": "question_id",
				"as":           "engagements",
			},
		},
		{
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
		{
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
		},
		// Continue building the pipeline as needed
	}
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

func (s *QuestionService) addFacetStageToPipeline(pipeline []bson.M, userID *primitive.ObjectID, answerStatus string) []bson.M {

	// if answerStatus includes "unattempted", then includeUnattempted is true
	includeUnattempted := strings.Contains(answerStatus, "unattempted")
	includeCorrect := strings.Contains(answerStatus, "correct")
	includeIncorrect := strings.Contains(answerStatus, "incorrect")
	includeOmitted := strings.Contains(answerStatus, "omitted")
	includeFlagged := strings.Contains(answerStatus, "flagged")

	facetStage := constructFacetStage(userID)

	if len(facetStage["$facet"].(bson.M)) > 0 {
		pipeline = append(pipeline, facetStage)

		setUnionFields := []interface{}{}

		if !includeUnattempted && !includeCorrect && !includeIncorrect && !includeOmitted && !includeFlagged {
			// If no statuses are included, include all statuses
			setUnionFields = append(setUnionFields, "$unattempted", "$correct", "$incorrect", "$omitted")
		} else {

			if includeUnattempted {
				setUnionFields = append(setUnionFields, "$unattempted")
			}
			if includeCorrect {
				setUnionFields = append(setUnionFields, "$correct")
			}
			if includeIncorrect {
				setUnionFields = append(setUnionFields, "$incorrect")
			}
			if includeOmitted {
				setUnionFields = append(setUnionFields, "$omitted")
			}

		}

		if len(setUnionFields) > 0 {
			pipeline = append(pipeline, bson.M{
				"$project": bson.M{
					"combined": bson.M{
						"$setUnion": setUnionFields,
					},
				},
			})

			pipeline = append(pipeline, bson.M{"$unwind": "$combined"})
			pipeline = append(pipeline, bson.M{"$replaceRoot": bson.M{"newRoot": "$combined"}})

			// Group by _id and take the first document in each group
			pipeline = append(pipeline, bson.M{"$group": bson.M{
				"_id": "$_id",
				"doc": bson.M{"$first": "$$ROOT"},
			}})

			pipeline = append(pipeline, bson.M{"$replaceRoot": bson.M{"newRoot": "$doc"}})
		}
	}

	return pipeline
}

func (s *QuestionService) addMatchAndSortStagesToPipeline(pipeline []bson.M, sortOption string, sortDirection string) []bson.M {
	// pipeline = append(pipeline, bson.M{"$match": filter})

	// Add sort stage to the pipeline
	sortStage := bson.M{}

	sortText := ""

	if sortOption == "topic" {
		sortText = "topic"
	} else if sortOption == "difficulty" {
		sortText = "difficultyLevel"
	} else if sortOption == "answerStatus" {
		sortText = "status"
	} else if sortOption == "attemptTime" {
		sortText = "first_attempt_time"
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

func (s *QuestionService) executePipeline(ctx context.Context, pipeline []bson.M) ([]*QuestionWithStatus, error) {
	cursor, err := s.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*QuestionWithStatus
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

func constructFacetStage(userID *primitive.ObjectID) bson.M {
	facets := bson.M{}

	if *userID == user.DefaultUserID {
		facets["unattempted"] = []bson.M{
			{
				"$match": bson.M{
					"engagements": bson.M{"$exists": true},
				},
			},
			{
				"$addFields": bson.M{
					"status": "unattempted",
				},
			},
		}
	} else {
		facets["unattempted"] = []bson.M{
			{
				"$match": bson.M{
					"engagements": bson.M{"$not": bson.M{"$elemMatch": bson.M{"user_id": userID}}},
				},
			},
			{
				"$addFields": bson.M{
					"status": "unattempted",
				},
			},
		}
	}

	facets["correct"] = []bson.M{
		{
			"$match": bson.M{
				"engagements": bson.M{
					"$elemMatch": bson.M{
						"user_id": userID,
						"status":  bson.M{"$eq": "correct"},
					},
				},
			},
		},
		{
			"$addFields": bson.M{
				"status": "correct",
				"first_attempt_time": bson.M{
					"$arrayElemAt": []interface{}{
						"$engagements.attempt_time",
						0,
					},
				},
			},
		},
	}

	facets["incorrect"] = []bson.M{
		{
			"$match": bson.M{
				"engagements": bson.M{
					"$elemMatch": bson.M{
						"user_id": userID,
						"status":  bson.M{"$eq": "incorrect"},
					},
				},
			},
		},
		{
			"$addFields": bson.M{
				"status": "incorrect",
				"first_attempt_time": bson.M{
					"$arrayElemAt": []interface{}{
						"$engagements.attempt_time",
						0,
					},
				},
			},
		},
	}

	facets["omitted"] = []bson.M{
		{
			"$match": bson.M{
				"engagements": bson.M{
					"$elemMatch": bson.M{
						"user_id": userID,
						"status":  bson.M{"$eq": "omitted"},
					},
				},
			},
		},
		{
			"$addFields": bson.M{
				"status": "omitted",
				"first_attempt_time": bson.M{
					"$arrayElemAt": []interface{}{
						"$engagements.attempt_time",
						0,
					},
				},
			},
		},
	}

	// facets["flagged"] = []bson.M{
	// 	{
	// 		"$match": bson.M{
	// 			"engagements": bson.M{
	// 				"$elemMatch": bson.M{
	// 					"user_id": userID,
	// 					"flagged": true,
	// 				},
	// 			},
	// 		},
	// 	},
	// 	{
	// 		"$addFields": bson.M{
	// 			"status": "flagged",
	// 		},
	// 	},
	// }

	return bson.M{"$facet": facets}
}

func shuffleQuestions(questions []*Question) []*Question {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(questions), func(i, j int) {
		questions[i], questions[j] = questions[j], questions[i]
	})
	return questions
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
