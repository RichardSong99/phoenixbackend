package videoengagement

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type VideoEngagementService struct {
	collection *mongo.Collection
}

func NewVideoEngagementService(client *mongo.Client) *VideoEngagementService {
	collection := client.Database("test").Collection("video_engagements")
	return &VideoEngagementService{
		collection: collection,
	}
}

func (s *VideoEngagementService) GetVideoEngagementByID(c context.Context, objid *primitive.ObjectID) (*VideoEngagement, error) {
	var videoEngagement VideoEngagement
	err := s.collection.FindOne(context.TODO(), bson.M{"_id": objid}).Decode(&videoEngagement)
	if err != nil {
		return nil, err
	}
	return &videoEngagement, nil
}

func (s *VideoEngagementService) LogVideoEngagement(c *gin.Context, engagement *VideoEngagement) (*mongo.UpdateResult, *mongo.InsertOneResult, error) {
	filter := bson.M{"user_id": *engagement.UserID, "video_id": *engagement.VideoID}
	update := bson.M{"$set": bson.M{"watched": engagement.Watched}}

	// Check if engagement already exists
	var result bson.M
	err := s.collection.FindOne(context.TODO(), filter).Decode(&result)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			// If no engagement found, insert a new one
			insertResult, err := s.collection.InsertOne(context.TODO(), bson.M{
				"user_id":  *engagement.UserID,
				"video_id": *engagement.VideoID,
				"watched":  engagement.Watched,
			})

			if err != nil {
				return nil, nil, err
			}

			return nil, insertResult, nil
		}

		return nil, nil, err
	}

	// If engagement found, update it
	updateResult, err := s.collection.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		return nil, nil, err
	}

	return updateResult, nil, nil
}
