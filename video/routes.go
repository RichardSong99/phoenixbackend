package video

import (
	"example/goserver/user"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/api/youtube/v3"
)

func RegisterRoutes(publicRouter *gin.RouterGroup, service *VideoService, youtubeService *youtube.Service) {
	// Add this line to create a new route for getVideos
	publicRouter.GET("/video", getVideo(service))
	publicRouter.GET("/videosbyid", getVideosByID(service))
	publicRouter.POST("/video", postVideo(service, youtubeService))
}

func getVideo(service *VideoService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add code to get videos from the service
		videoObjId := c.DefaultQuery("videoObjId", "000000000000000000000000")
		videObjIdOID, err := primitive.ObjectIDFromHex(videoObjId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid video ID"})
			return
		}

		video, err := service.GetVideo(videObjIdOID)
		// Continue with the rest of the code

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, video)
	}
}

func postVideo(service *VideoService, youtubeService *youtube.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var video Video

		if err := c.ShouldBindJSON(&video); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Fetch video details from YouTube's API
		call := youtubeService.Videos.List([]string{"snippet"}).Id(video.VideoID)
		response, err := call.Do()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Update video details
		if len(response.Items) > 0 {
			item := response.Items[0]
			video.Title = item.Snippet.Title
			video.Description = item.Snippet.Description
			video.Thumbnail = item.Snippet.Thumbnails.Default.Url
		}

		id, err := service.PostVideo(c, &video)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Video posted successfully", "id": id})
	}
}

func getVideosByID(service *VideoService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		var userIDObj *primitive.ObjectID

		if exists {
			// Convert userID to *primitive.ObjectID
			userIDObjTemp, err := primitive.ObjectIDFromHex(userID.(string))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
				return
			}
			userIDObj = &userIDObjTemp
		} else {
			// Create a static userID for unlogged user
			defaultUserID := user.DefaultUserID
			userIDObj = &defaultUserID
		}

		err := error(nil) // Declare the "err" variable
		videoids := c.QueryArray("ids")
		videoIDsObj := make([]primitive.ObjectID, len(videoids))
		for i, id := range videoids {
			videoIDsObj[i], err = primitive.ObjectIDFromHex(id)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid video ID"})
				return
			}
		}

		videos, err := service.GetVideosByID(c, videoIDsObj, userIDObj)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		numTotal := len(videos)
		numWatched := 0
		numFlagged := 0
		for _, video := range videos {
			if video.Watched {
				numWatched++
			}
			if video.Flagged {
				numFlagged++
			}
		}

		percentWatched := 0.0
		if numTotal > 0 {
			percentWatched = float64(numWatched) / float64(numTotal)
		}

		c.JSON(http.StatusOK, gin.H{"videos": videos, "numTotal": numTotal, "numWatched": numWatched, "numFlagged": numFlagged, "percentWatched": percentWatched})
	}
}
