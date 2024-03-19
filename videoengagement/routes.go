package videoengagement

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func RegisterRoutes(publicRouter *gin.RouterGroup, s *VideoEngagementService) {
	publicRouter.GET("/videoengagement/:id", s.getVideoEngagementByID)
	publicRouter.POST("/videoengagement", s.postVideoEngagement)

}

func (s *VideoEngagementService) getVideoEngagementByID(c *gin.Context) {
	id := c.Param("id")

	// convert id to primitive.ObjectID
	idObj, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	engagement, err := s.GetVideoEngagementByID(c, &idObj)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, engagement)
}

func (s *VideoEngagementService) postVideoEngagement(c *gin.Context) {
	// parameters: videoId and userId
	userID, exists := c.Get("userID")
	var userIDObj *primitive.ObjectID

	videoIdStr := c.DefaultQuery("videoId", "000000000000000000000000")
	videoID, err := primitive.ObjectIDFromHex(videoIdStr)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid video ID"})
		return
	}

	watched := c.Query("watched")

	if exists {
		// Convert userID to *primitive.ObjectID
		userIDObjTemp, err := primitive.ObjectIDFromHex(userID.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
			return
		}
		userIDObj = &userIDObjTemp

		if watched == "true" {
			updateResult, insertResult, err := s.LogVideoEngagement(c, &VideoEngagement{UserID: userIDObj, VideoID: &videoID, Watched: true})

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			if updateResult != nil {
				fmt.Printf("Updated document %v\n", updateResult)
			}

			if insertResult != nil {
				fmt.Printf("Inserted document with ID %v\n", insertResult.InsertedID)
			}

			c.JSON(http.StatusOK, gin.H{"message": "Engagement logged successfully"})
		}
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Engagement not logged because user not logged in"})

	}

	// return message saying engagement not logged because user not logged in...

}
