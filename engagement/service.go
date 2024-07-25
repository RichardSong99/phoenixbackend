package engagement

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type EngagementService struct {
	collection *mongo.Collection
}

// NewEngagementService creates a new engagement service
func NewEngagementService(client *mongo.Client) *EngagementService {
	collection := client.Database("test").Collection("engagements")
	return &EngagementService{
		collection: collection,
	}
}

func (s *EngagementService) GetEngagementCollection() *mongo.Collection {
	return s.collection
}

func (s *EngagementService) GetEngagementByUserAndQuestionID(ctx context.Context, userID, questionID *primitive.ObjectID) (*Engagement, error) {
	var engagement Engagement
	err := s.collection.FindOne(ctx, bson.M{"user_id": userID, "question_id": questionID}).Decode(&engagement)
	if err != nil {
		return nil, err
	}
	return &engagement, nil
}

// LogEngagement logs an engagement to the database
func (es *EngagementService) LogEngagement(ctx context.Context, engagement *Engagement) (string, error) {
	// There can only be one engagement with the same userID and questionID.
	// Try to find an engagement with the same userID and questionID.
	// If found, update the existing engagement with the new attempt.
	// If not found, insert a new engagement.

	filter := bson.M{"user_id": engagement.UserID, "question_id": engagement.QuestionID}

	// Define update operation
	update := bson.M{"$set": engagement}

	// Options for the update operation
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	// Find and update existing engagement
	var updatedEngagement Engagement
	err := es.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedEngagement)
	if err != nil {
		return "", err
	}

	return updatedEngagement.ID.Hex(), nil
}

// GetEngagementByID retrieves an engagement from the database by ID
func (es *EngagementService) GetEngagementByID(ctx context.Context, engagementID primitive.ObjectID) (*Engagement, error) {
	var engagement Engagement
	err := es.collection.FindOne(ctx, bson.M{"_id": engagementID}).Decode(&engagement)
	if err != nil {
		return nil, err
	}

	return &engagement, nil
}

// GetEngagementsByID
func (es *EngagementService) GetEngagementsByID(ctx context.Context, engagementIDs []primitive.ObjectID) ([]*Engagement, error) {
	cursor, err := es.collection.Find(ctx, bson.M{"_id": bson.M{"$in": engagementIDs}})
	if err != nil {
		return nil, err
	}

	var engagements []*Engagement
	if err = cursor.All(ctx, &engagements); err != nil {
		return nil, err
	}

	return engagements, nil
}

// UpdateEngagement updates an engagement in the database
func (es *EngagementService) UpdateEngagement(ctx context.Context, id string, update bson.M) (*mongo.UpdateResult, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	result, err := es.collection.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": update}, options.Update().SetUpsert(true))
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *EngagementService) GetAttemptedQuestionIDs(ctx context.Context, userID *primitive.ObjectID) ([]*primitive.ObjectID, error) {
	cursor, err := s.GetEngagementCollection().Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var engagements []*Engagement
	if err = cursor.All(ctx, &engagements); err != nil {
		return nil, err
	}
	var ids []*primitive.ObjectID
	for _, engagement := range engagements {
		ids = append(ids, engagement.QuestionID)
	}
	return ids, nil
}
