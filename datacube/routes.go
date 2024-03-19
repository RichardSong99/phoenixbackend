package datacube

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func RegisterRoutes(publicRouter *gin.RouterGroup, authRouter *gin.RouterGroup, dataCubeService *DataCubeService) {
	// Existing code...

	// Add this line to create a new route for getDatacube
	publicRouter.GET("/datacube", getDatacube(dataCubeService))

	// Existing code...
}

func getDatacube(service *DataCubeService) gin.HandlerFunc {
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
			userIDObjTemp, err := primitive.ObjectIDFromHex("000000000000000000000000")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating ObjectID for unlogged user"})
				return
			}
			userIDObj = &userIDObjTemp
		}

		compute := c.DefaultQuery("compute", "false")

		var dataCube *DataCube
		var err error

		if compute == "true" {
			dataCube, err = service.ComputeDataCube(userIDObj)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		} else {
			dataCube, err = service.GetDataCube(userIDObj)

			// if err != nil {
			// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			// 	return
			// }

			if dataCube == nil {
				dataCube, err = service.ComputeDataCube(userIDObj)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
			}
		}

		c.JSON(http.StatusOK, dataCube)
	}
}
