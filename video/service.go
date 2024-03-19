package video

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type VideoService struct {
	collection *mongo.Collection
}

func NewVideoService(client *mongo.Client) *VideoService {
	collection := client.Database("test").Collection("videos")
	return &VideoService{
		collection: collection,
	}
}

func (s *VideoService) GetVideo(videoObjIdOID primitive.ObjectID) (Video, error) {
	var video Video

	err := s.collection.FindOne(context.TODO(), bson.M{"_id": videoObjIdOID}).Decode(&video)
	if err != nil {
		return Video{}, err
	}

	return video, nil
}

func (s *VideoService) PostVideo(c context.Context, video *Video) (primitive.ObjectID, error) {
	res, err := s.collection.InsertOne(c, video)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return res.InsertedID.(primitive.ObjectID), nil
}

func (s *VideoService) GetVideosByID(c context.Context, videoIDs []primitive.ObjectID, userID *primitive.ObjectID) ([]*VideoWithStatus, error) {
	pipeline := []bson.M{{"$match": bson.M{"_id": bson.M{"$in": videoIDs}}}}

	if userID != nil {
		pipeline = append(pipeline, bson.M{"$lookup": bson.M{
			"from":         "video_engagements",
			"localField":   "_id",
			"foreignField": "video_id",
			"as":           "engagements",
		}})

		// add a watched field based on if there is an engagement with the matching userId and watched = true
		pipeline = append(pipeline, bson.M{"$addFields": bson.M{
			"Watched": bson.M{"$cond": []interface{}{
				bson.M{"$eq": []interface{}{bson.M{"$size": "$engagements"}, 0}},
				false,
				bson.M{"$cond": []interface{}{
					bson.M{"$gt": []interface{}{
						bson.M{"$size": bson.M{"$filter": bson.M{
							"input": "$engagements",
							"as":    "engagement",
							"cond":  bson.M{"$and": []bson.M{{"$eq": []interface{}{"$$engagement.user_id", userID}}, {"$eq": []interface{}{"$$engagement.watched", true}}}},
						}}},
						0,
					}},
					true,
					false,
				}},
			}},
		}})

		// add a flagged field based on if there is an engagement with the matching userId and flagged = true
		pipeline = append(pipeline, bson.M{"$addFields": bson.M{
			"Flagged": bson.M{"$cond": []interface{}{
				bson.M{"$gt": []interface{}{
					bson.M{"$size": bson.M{"$filter": bson.M{
						"input": "$engagements",
						"as":    "engagement",
						"cond":  bson.M{"$and": []bson.M{{"$eq": []interface{}{"$$engagement.user_id", userID}}, {"$eq": []interface{}{"$$engagement.flagged", true}}}},
					}}},
					0,
				}},
				true,
				false,
			}},
		}})
	}

	// Add a project stage to add the watched and flagged fields
	pipeline = append(pipeline, bson.M{"$project": bson.M{
		"video":   "$$ROOT",
		"Watched": "$Watched",
		"Flagged": "$Flagged",
	}})

	cursor, err := s.collection.Aggregate(c, pipeline)
	if err != nil {
		return nil, err
	}

	var videos []*VideoWithStatus
	err = cursor.All(c, &videos)
	if err != nil {
		return nil, err
	}

	return videos, nil
}
