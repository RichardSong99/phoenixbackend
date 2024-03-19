package lessons

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(publicRouter *gin.RouterGroup, lessonService *LessonService, courseService *CourseService) {
	publicRouter.POST("/lesson", AddLessonHandler(lessonService))
	publicRouter.GET("/lesson/:id", GetLessonHandler(lessonService))
	publicRouter.PUT("/lesson/:id", UpdateLessonHandler(lessonService))
	publicRouter.POST("/course", AddCourseHandler(courseService))
	publicRouter.GET("/course/:id", GetCourseHandler(courseService))
}

func AddLessonHandler(service *LessonService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var lesson Lesson
		if err := c.ShouldBindJSON(&lesson); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		id, err := service.AddLesson(c, &lesson)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Lesson added successfully", "id": id})
	}
}

func GetLessonHandler(service *LessonService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		lesson, err := service.GetLessonByID(c, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, lesson)
	}
}

func UpdateLessonHandler(service *LessonService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var lesson Lesson

		if err := c.ShouldBindJSON(&lesson); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := service.UpdateLesson(c, id, &lesson)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Lesson updated successfully", "result": result})
	}
}
