package parameterdata

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(publicRouter *gin.RouterGroup, authRouter *gin.RouterGroup) {
	// Add this line to create a new route for getDatacube
	publicRouter.GET("/topiclist", getTopicList())
	publicRouter.GET("/lessonmodule", getLessonModule())
	publicRouter.GET("/practicemodule", getPracticeModule())
	publicRouter.GET("/testrepresentation", getTestRepresentation())
	// Existing code...
}

func getTestRepresentation() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add code to get the test representation
		name := c.Query("name")
		var testRepresentation *TestRepresentation
		for _, representation := range Tests {
			if representation.Name == name {
				testRepresentation = representation
				break
			}
		}

		if testRepresentation == nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "test representation not found"})
			return
		}

		c.JSON(http.StatusOK, testRepresentation)

	}
}

func getTopicList() gin.HandlerFunc {
	return func(c *gin.Context) {

		subject := c.DefaultQuery("subject", "math")

		var topicList []*Topic

		if subject == "math" {
			topicList = MathTopicsList
		} else {
			topicList = ReadingTopicsList
		}

		c.JSON(http.StatusOK, topicList)
	}
}

func getLessonModule() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add code to get the lesson module
		name := c.Query("name")
		var lessonModule *LessonModule
		for _, module := range LessonModules {
			if module.Name == name {
				lessonModule = module
				break
			}
		}

		if lessonModule == nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "lesson module not found"})
			return
		}

		c.JSON(http.StatusOK, lessonModule)

	}
}

func getPracticeModule() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add code to get the practice module
		name := c.Query("name")
		var practiceModule *PracticeModule
		for _, module := range PracticeModules {
			if module.Name == name {
				practiceModule = module
				break
			}
		}

		if practiceModule == nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "practice module not found"})
			return
		}

		c.JSON(http.StatusOK, practiceModule)

	}
}

// func getContentList() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		topic := c.DefaultQuery("topic", "Linear equations in 1 variable")
// 		var contentList *ContentList

// 		for _, list := range ListOfQuestionLists {
// 			if list.Name == topic {
// 				contentList = list
// 				break
// 			}
// 		}

// 		if contentList == nil {
// 			c.JSON(http.StatusNotFound, gin.H{"message": "content list not found"})
// 			return
// 		}

// 		c.JSON(http.StatusOK, contentList)
// 	}
// }
