package engagement

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func RegisterRoutes(r *gin.Engine, engagementService *EngagementService) {
	r.POST("/engagement", LogEngagementHandler(engagementService))
	r.POST("/engagements", LogEngagementsHandler(engagementService)) // New route for logging multiple engagements

	r.GET("/engagement/:id", GetEngagementByIDHandler(engagementService))
	r.GET("/engagement", GetEngagementHandler(engagementService)) // New route for getting engagement by ID
	r.PATCH("/engagement/:id", UpdateEngagementHandler(engagementService))

}

func GetEngagementHandler(service *EngagementService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")

		if !exists {
			c.JSON(http.StatusOK, gin.H{"message": "Engagement not logged because user not logged in"})
			return
		}

		// Convert the UserID to an ObjectID
		userIDObjID, err := primitive.ObjectIDFromHex(userID.(string))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		questionID := c.Query("questionID")
		questionIDObjID, err := primitive.ObjectIDFromHex(questionID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid question ID"})
			return
		}

		engagement, err := service.GetEngagementByUserAndQuestionID(c, &userIDObjID, &questionIDObjID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, engagement)
	}
}

// LogEngagementHandler is a gin HandlerFunc that logs an engagement
func LogEngagementHandler(service *EngagementService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var engagement Engagement

		if err := c.ShouldBindJSON(&engagement); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Attempt to get user ID from context
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusOK, gin.H{"message": "Engagement not logged because user not logged in"})
			return
		}

		// Convert the UserID to an ObjectID
		userIDObjID, err := primitive.ObjectIDFromHex(userID.(string))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		// Set the UserID in the engagement
		engagement.UserID = &userIDObjID

		id, err := service.LogEngagement(c, &engagement)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Engagement logged successfully", "id": id})
	}
}

// LogEngagementsHandler is a gin HandlerFunc that logs multiple engagements
func LogEngagementsHandler(service *EngagementService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var engagements []Engagement

		if err := c.ShouldBindJSON(&engagements); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get the UserID from the context
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusOK, gin.H{"message": "Engagements not logged because user not logged in"})
			return
		}

		// Convert the UserID to an ObjectID
		userIDObjID, err := primitive.ObjectIDFromHex(userID.(string))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		// Set the UserID in each engagement
		for i := range engagements {
			engagements[i].UserID = &userIDObjID
			// use the LogEngagement Function to log each engagement
			id, err := service.LogEngagement(c, &engagements[i])
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			fmt.Println("Engagement logged successfully", id)
		}
		c.JSON(http.StatusOK, gin.H{"message": "Engagements logged successfully"})
	}
}

// GetEngagementByIDHandler is a gin HandlerFunc that gets an engagement by ID
func GetEngagementByIDHandler(service *EngagementService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		// Convert id to primitive.ObjectID
		idObj, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
			return
		}

		engagement, err := service.GetEngagementByID(c, idObj)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, engagement)
	}
}

// ConvertJSONToBSONFields converts JSON field names to BSON field names based on struct tags.
func ConvertJSONToBSONFields(jsonFields map[string]interface{}, model interface{}) (bson.M, error) {
	t := reflect.TypeOf(model)
	fieldNameMap := make(map[string]string)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		bsonTag := field.Tag.Get("bson")
		if jsonTag != "" && bsonTag != "" {
			jsonFieldName := strings.Split(jsonTag, ",")[0]
			bsonFieldName := strings.Split(bsonTag, ",")[0]
			fieldNameMap[jsonFieldName] = bsonFieldName
		}
	}

	update := bson.M{}
	for jsonField, value := range jsonFields {
		bsonField, ok := fieldNameMap[jsonField]
		if !ok {
			return nil, fmt.Errorf("Invalid field name: %s", jsonField)
		}
		update[bsonField] = value
	}

	return update, nil
}

func UpdateEngagementHandler(service *EngagementService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var jsonFields map[string]interface{}

		if err := c.ShouldBindJSON(&jsonFields); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		update, err := ConvertJSONToBSONFields(jsonFields, Engagement{})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := service.UpdateEngagement(c, id, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Engagement updated successfully", "result": result})
	}
}
