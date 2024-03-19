package lessons

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Similar handlers for Course
func AddCourseHandler(service *CourseService) gin.HandlerFunc {
	// Similar to AddLessonHandler
	return func(c *gin.Context) {
		var course Course
		if err := c.ShouldBindJSON(&course); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		id, err := service.AddCourse(c, &course)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Lesson added successfully", "id": id})
	}
}

func GetCourseHandler(service *CourseService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		course, err := service.GetCourseByID(c, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, course)
	}
}

func AddLessonToCourseHandler(service *CourseService) gin.HandlerFunc {
	return func(c *gin.Context) {
		courseID := c.Param("courseID")
		lessonID := c.Param("lessonID")

		if courseID == "" || lessonID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Course ID and Lesson ID must be specified"})
			return
		}

		indexStr := c.DefaultQuery("index", "-1")
		index, err := strconv.Atoi(indexStr)
		if err != nil {
			index = -1 // If index is not a valid integer, add the lesson to the end of the slice
		}

		_, err = service.AddLessonToCourse(c, courseID, lessonID, index)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Lesson added to course successfully"})
	}
}
