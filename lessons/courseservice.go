package lessons

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CourseService struct {
	collection *mongo.Collection
}

func NewCourseService(client *mongo.Client) *CourseService {
	collection := client.Database("test").Collection("courses")
	return &CourseService{collection: collection}
}

func (c *CourseService) AddCourse(ctx context.Context, course *Course) (*Course, error) {
	_, err := c.collection.InsertOne(ctx, course)
	if err != nil {
		return nil, err
	}
	return course, nil
}

func (c *CourseService) GetCourseByID(ctx context.Context, id string) (*Course, error) {
	var course Course

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	err = c.collection.FindOne(ctx, bson.M{"_id": oid}).Decode(&course)
	if err != nil {
		return nil, err
	}
	return &course, nil
}

func (c *CourseService) AddLessonToCourse(ctx context.Context, courseID string, lessonID string, index int) (*Course, error) {
	course, err := c.GetCourseByID(ctx, courseID)
	if err != nil {
		return nil, err
	}

	lessonOID, err := primitive.ObjectIDFromHex(lessonID)
	if err != nil {
		return nil, err
	}

	if index < 0 || index > len(course.LessonIDs) {
		course.LessonIDs = append(course.LessonIDs, &lessonOID)
	} else {
		// Insert lessonOID at the specified index
		course.LessonIDs = append(course.LessonIDs[:index], append([]*primitive.ObjectID{&lessonOID}, course.LessonIDs[index:]...)...)
	}

	_, err = c.collection.ReplaceOne(ctx, bson.M{"_id": course.ID}, course)
	if err != nil {
		return nil, err
	}
	return course, nil
}
