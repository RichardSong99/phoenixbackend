package question

import (
	"example/goserver/user"
	"fmt"
	"net/http"
	"strconv"
	"time"

	// replace with your project path
	// replace with your project path

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var questionService *QuestionService

// RegisterRoutes registers the question routes
func RegisterRoutes(publicRouter *gin.RouterGroup, authRouter *gin.RouterGroup, service *QuestionService, userService *user.UserService) {
	questionService = service

	// Public route, accessible to both authenticated and unauthenticated users
	// publicRouter.GET("/questions/masked", getMaskedQuestions(userService))

	// Authenticated routes, only accessible to authenticated users
	publicRouter.GET("/question/:id", getQuestion(userService))
	publicRouter.GET("/questions", getQuestions(userService, questionService))
	publicRouter.GET("/questions/data", getQuestionStatistics(questionService))
	publicRouter.GET("/questionsbyid", getQuestionsByIDHandler(questionService))
	publicRouter.PUT("/questions", updateAllQuestions(questionService)) // Add this line

	// Assuming these are admin-only routes, you can keep them under authenticated routes
	// and add further authorization checks as needed
	publicRouter.POST("/questions", createQuestion)
	publicRouter.PUT("/questions/:id", updateQuestion)
	publicRouter.DELETE("/questions/:id", deleteQuestion)
}

// createQuestion handles the POST /questions route
func createQuestion(c *gin.Context) {
	var question Question
	if err := c.ShouldBindJSON(&question); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set the creation and last edited dates to the current time
	currentTime := time.Now()
	question.CreationDate = currentTime
	question.LastEditedDate = currentTime

	result, err := questionService.CreateQuestion(c, &question)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// getQuestion handles the GET /questions/:id route
func getQuestion(userService *user.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}

		// Attempt to get user ID from JWT token
		userTier := userService.GetUserTier(c)

		fmt.Println("userTier", userTier)

		question, err := questionService.GetQuestion(c, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Dereference question.AccessOption before comparing
		if question.AccessOption != nil && *question.AccessOption == "paid" && userTier != "paid" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}

		c.JSON(http.StatusOK, question)
	}
}

func getQuestionsByIDHandler(questionService *QuestionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		questionIDs := c.QueryArray("ids")

		// convert into array of object ids
		var questionIDsObj []primitive.ObjectID
		for _, id := range questionIDs {
			idObj, err := primitive.ObjectIDFromHex(id)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
				return
			}
			questionIDsObj = append(questionIDsObj, idObj)
		}

		questions, err := questionService.GetQuestionsByID(c, questionIDsObj)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, questions)
	}
}

func getQuestionsByIDOld(questionService *QuestionService) gin.HandlerFunc {
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
			// Create a default userIDObj with a value of "0000..."
			defaultUserID := user.DefaultUserID
			userIDObj = &defaultUserID
		}

		questionIDs := c.QueryArray("ids")

		// convert into array of object ids
		var questionIDsObj []primitive.ObjectID
		for _, id := range questionIDs {
			idObj, err := primitive.ObjectIDFromHex(id)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
				return
			}
			questionIDsObj = append(questionIDsObj, idObj)
		}

		questions, err := questionService.GetQuestionsByIDOld(c, questionIDsObj, userIDObj)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// return object with: array of questions: number of total questions, number of answered questions (i.e., status != null), and number of correct questions (i.e., status == "correct")
		numTotal := len(questions)
		numAnswered := 0
		numCorrect := 0
		for _, q := range questions {
			if q.Status != nil {
				if *q.Status != "unattempted" {
					numAnswered++
				}
				if *q.Status == "correct" {
					numCorrect++
				}
			}
		}

		percentAnswered := 0.0

		if numTotal != 0 {
			percentAnswered = float64(numAnswered) / float64(numTotal)
		}

		percentCorrect := 0.0
		if numAnswered != 0 {
			percentCorrect = float64(numCorrect) / float64(numAnswered)
		}

		c.JSON(http.StatusOK, gin.H{
			"questions":       questions,
			"numTotal":        numTotal,
			"numAnswered":     numAnswered,
			"numCorrect":      numCorrect,
			"percentAnswered": percentAnswered,
			"percentCorrect":  percentCorrect,
		})

	}
}

func getQuestionStatistics(questionService *QuestionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// get the user ID from the context
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
			// Create a default userIDObj with a value of "0000..."
			defaultUserID := user.DefaultUserID
			userIDObj = &defaultUserID
		}

		dataQuery := c.DefaultQuery("data", "difficulty")

		var statistics interface{}
		var err error

		switch dataQuery {
		case "difficulty":
			statistics, err = questionService.GetDifficultyStatistics(c, userIDObj)
		case "status":
			statistics, err = questionService.GetStatusStatistics(c, userIDObj)
		case "combined":
			statistics, err = questionService.GetCombinedCubeStatistics(c, userIDObj)
		case "time":
			statistics, err = questionService.GetTimeStatistics(c, userIDObj)
		// Add more cases as needed...
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data query"})
			return
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Print out the statistics map for debugging

		c.JSON(http.StatusOK, statistics)
	}
}

func getQuestions(userService *user.UserService, questionService *QuestionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		topic := c.Query("topic")
		difficulty := c.Query("difficulty")
		answerStatus := c.Query("answerStatus")
		answerType := c.Query("answerType")

		subject := c.Query("subject")

		sortOption := c.Query("sortOption")
		sortDirection := c.Query("sortDirection")

		// Get page and pageSize parameters from query string
		pageStr := c.DefaultQuery("page", "1")
		pageSizeStr := c.DefaultQuery("pageSize", "10")

		page, err := strconv.ParseInt(pageStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
			return
		}

		pageSize, err := strconv.ParseInt(pageSizeStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page size"})
			return
		}

		// Calculate the number of documents to skip
		skip := (page - 1) * pageSize

		// Attempt to get user ID from JWT token
		userTier := userService.GetUserTier(c)

		// Get the user's attempted question IDs
		// Attempt to get user ID from context
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
		}

		questions, totalQuestions, err := questionService.GetQuestions(c, difficulty, topic, answerStatus, answerType, skip, pageSize, userTier, userIDObj, subject, sortOption, sortDirection)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		lastPage := totalQuestions / pageSize
		if totalQuestions%pageSize > 0 {
			lastPage++
		}

		c.JSON(http.StatusOK, gin.H{
			"currentPage":    page,
			"lastPage":       lastPage,
			"totalQuestions": totalQuestions,
			"data":           questions,
		})
	}
}

func updateAllQuestions(questionService *QuestionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		subject := c.Query("subject")

		update := bson.M{"subject": subject}
		result, err := questionService.UpdateAllQuestions(c, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"matchedDocuments":  result.MatchedCount,
			"modifiedDocuments": result.ModifiedCount,
		})
	}
}

// updateQuestion handles the PUT /questions/:id route
func updateQuestion(c *gin.Context) {
	// Parse the ID
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// Parse the request body
	var questionUpdate map[string]interface{}
	if err := c.ShouldBindJSON(&questionUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Normalize the field names to match the existing schema
	normalizedUpdate := normalizeFieldNames(questionUpdate)

	// Set the LastEditedDate to the current date and time
	normalizedUpdate["last_edited_date"] = time.Now().UTC()

	// If CreationDate is not provided, fetch the existing question to get its CreationDate
	if _, ok := normalizedUpdate["creation_date"]; !ok {
		existingQuestion, err := questionService.GetQuestionByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		normalizedUpdate["creation_date"] = existingQuestion.CreationDate
	}

	// Update the question
	result, err := questionService.UpdateQuestion(c.Request.Context(), id, normalizedUpdate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the updated question
	c.JSON(http.StatusOK, result)
}

// normalizeFieldNames converts the incoming JSON field names to match the existing schema
func normalizeFieldNames(update map[string]interface{}) bson.M {
	normalized := bson.M{}

	for key, value := range update {
		switch key {
		case "Prompt":
			normalized["prompt"] = value
		case "AnswerType":
			normalized["answer_type"] = value
		case "AnswerChoices":
			normalized["answer_choices"] = value
		case "CorrectAnswerMultiple":
			normalized["correct_answer_multiple"] = value
		case "CorrectAnswerFree":
			normalized["correct_answer_free"] = value
		case "Text":
			normalized["text"] = value
		case "Subject":
			normalized["subject"] = value
		case "Topic":
			normalized["topic"] = value
		case "Difficulty":
			normalized["difficulty"] = value
		case "AccessOption":
			normalized["access_option"] = value
		case "Explanation":
			normalized["explanation"] = value
		case "Images":
			normalized["images"] = value
		case "CreationDate":
			normalized["creation_date"] = value
		case "LastEditedDate":
			normalized["last_edited_date"] = value
		default:
			normalized[key] = value
		}
	}

	return normalized
}

// deleteQuestion handles the DELETE /questions/:id route
func deleteQuestion(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	result, err := questionService.DeleteQuestion(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
