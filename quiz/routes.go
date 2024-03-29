package quiz

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"example/goserver/engagement"
	"example/goserver/question"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func RegisterRoutes(publicRouter *gin.RouterGroup, service *QuizService, questionService *question.QuestionService, engagementService *engagement.EngagementService) {
	// Add this line to create a new route for getQuiz
	publicRouter.POST("/quiz", initializeQuiz(service))
	publicRouter.PATCH("/quizzes/:quizID/engagements/:engagementID", updateQuiz(service))
	publicRouter.GET("/quiz", getQuiz(service))
	publicRouter.GET("/quiz/:id/underlying", getQuizUnderlying(service, questionService, engagementService))
	publicRouter.GET("/quizzes", getQuizzesForUser(service))
}

func initializeQuiz(service *QuizService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add code to initialize a new quiz
		// the input will be an array of questionIDs
		// the output will be a quizID

		var requestData struct {
			QuestionIDList []string `json:"QuestionIDList"`
			Type           *string  `json:"Type"`
			Name           *string  `json:"Name"`
		}

		if err := c.ShouldBindJSON(&requestData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userID, exists := c.Get("userID")
		var userIDObj *primitive.ObjectID

		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User not logged in"})
			return
		}

		// Convert userID to *primitive.ObjectID
		userIDObjTemp, err := primitive.ObjectIDFromHex(userID.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
			return
		}
		userIDObj = &userIDObjTemp

		// fmt.Println("userID", userIDObj)

		//convert questionIDs to ObjectIDs
		questionIDsObjIDs := make([]primitive.ObjectID, len(requestData.QuestionIDList))
		for i, id := range requestData.QuestionIDList {
			questionIDObjID, err := primitive.ObjectIDFromHex(id)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid question ID"})
				return
			}
			questionIDsObjIDs[i] = questionIDObjID
		}

		if requestData.Type == nil {
			requestData.Type = new(string)
		}

		if requestData.Name == nil {
			name := time.Now().Format("2006-01-02 15:04:05")
			requestData.Name = &name
		}

		quizID, err := service.InitializeQuiz(c, questionIDsObjIDs, *userIDObj, requestData.Type, requestData.Name)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"quizID": quizID})
	}

}

func updateQuiz(service *QuizService) gin.HandlerFunc {
	// request body should contain the quizID and an engagement ID
	return func(c *gin.Context) {
		// Add code to update a quiz
		// the input will be a quizID and an engagementID
		// the output will be a quizID

		// userID, exists := c.Get("userID")
		// var userIDObj *primitive.ObjectID

		// if exists {
		// 	// Convert userID to *primitive.ObjectID
		// 	userIDObjTemp, err := primitive.ObjectIDFromHex(userID.(string))
		// 	if err != nil {
		// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		// 		return
		// 	}
		// 	userIDObj = &userIDObjTemp

		engagementID, err := primitive.ObjectIDFromHex(c.Param("engagementID"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid engagement ID"})
			return
		}

		quizID, err := primitive.ObjectIDFromHex(c.Param("quizID"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid quiz ID"})
			return
		}

		quizID, err = service.UpdateQuiz(c, quizID, engagementID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"quizID": quizID})

	}
}

func getQuiz(service *QuizService) gin.HandlerFunc {

	return func(c *gin.Context) {
		// Add code to get a quiz
		// the input will be a quizID
		// the output will be a quiz

		fmt.Println("in here")

		quizID := c.Query("id")
		// convert quizID to ObjectID

		quizName := c.Query("name")

		if quizID == "" && quizName == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid quiz ID or name"})
			return
		}

		if quizID != "" {
			quizIDObjID, err := primitive.ObjectIDFromHex(quizID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid quiz ID"})
				return
			}
			quiz, err := service.GetQuiz(c, quizIDObjID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, quiz)
			return
		}

		userID, exists := c.Get("userID")
		var userIDObj *primitive.ObjectID

		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User not logged in"})
			return
		}

		// Convert userID to *primitive.ObjectID
		userIDObjTemp, err := primitive.ObjectIDFromHex(userID.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
			return
		}
		userIDObj = &userIDObjTemp

		if quizName != "" {
			// need to decode the encoded quizName
			originalQuizName, err := url.QueryUnescape(quizName)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid quiz name"})
				return
			}

			quiz, err := service.GetQuizByName(c, originalQuizName, *userIDObj)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, quiz)
		}
	}
}

func getQuizUnderlying(service *QuizService, questionService *question.QuestionService, engagementService *engagement.EngagementService) gin.HandlerFunc {
	return func(c *gin.Context) {
		quizID, err := primitive.ObjectIDFromHex(c.Param("id"))

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid quiz ID"})
			return
		}

		result, err := GetQuizUnderlying(c, service, questionService, engagementService, quizID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func GetQuizUnderlying(ctx context.Context, service *QuizService, questionService *question.QuestionService, engagementService *engagement.EngagementService, quizID primitive.ObjectID) (*QuizResult, error) {
	quiz, err := service.GetQuiz(ctx, quizID)
	if err != nil {
		return nil, err
	}

	questions := make([]question.Question, len(quiz.QuestionIDList))

	for i, questionID := range quiz.QuestionIDList {
		question, err := questionService.GetQuestionByID(ctx, questionID)
		if err != nil {
			return nil, err
		}
		questions[i] = *question
	}

	questionEngagementCombos := make([]QuestionEngagementCombo, len(questions))
	for i, question := range questions {
		localQuestion := question
		found := false
		for _, engagementID := range quiz.EngagementIDList {
			localEngagementID := engagementID
			engagement, err := engagementService.GetEngagementByID(ctx, localEngagementID)
			if err != nil {
				return nil, err
			}

			if engagement.QuestionID.Hex() == localQuestion.ID.Hex() {
				questionEngagementCombos[i] = QuestionEngagementCombo{Question: &localQuestion, Engagement: engagement}
				found = true
				break
			}
		}

		if !found {
			questionEngagementCombos[i] = QuestionEngagementCombo{Question: &localQuestion, Engagement: nil}
		}
	}

	numTotal := len(questions)
	numAnswered := 0
	numCorrect := 0
	numIncorrect := 0
	numOmitted := 0
	numUnattempted := 0
	percentAnswered := 0.0
	percentCorrect := 0.0

	for _, QuestionEngagementCombo := range questionEngagementCombos {
		engagement := QuestionEngagementCombo.Engagement
		if engagement != nil {
			if *engagement.Status == "correct" {
				numCorrect++
				numAnswered++
			} else if *engagement.Status == "incorrect" {
				numIncorrect++
				numAnswered++
			} else if *engagement.Status == "omitted" {
				numOmitted++
				numAnswered++
			}
		}
	}

	numUnattempted = numTotal - numAnswered
	if numTotal != 0 {
		percentAnswered = float64(numAnswered) / float64(numTotal) * 100
	}

	if numAnswered != 0 {
		percentCorrect = float64(numCorrect) / float64(numAnswered) * 100
	}

	return &QuizResult{
		Quiz:            quiz,
		Questions:       questionEngagementCombos,
		NumTotal:        numTotal,
		NumAnswered:     numAnswered,
		NumCorrect:      numCorrect,
		NumIncorrect:    numIncorrect,
		NumOmitted:      numOmitted,
		NumUnattempted:  numUnattempted,
		PercentAnswered: percentAnswered,
		PercentCorrect:  percentCorrect,
	}, nil
}

func getQuizzesForUser(service *QuizService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add code to get all quizzes for a user
		// the input will be a userID
		// the output will be an array of quizzes

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

			quizzes, err := service.GetQuizzesForUser(c, *userIDObj)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, quizzes)
		} else {
			c.JSON(http.StatusOK, gin.H{"error": "User not logged in"})
			return
		}
	}
}
