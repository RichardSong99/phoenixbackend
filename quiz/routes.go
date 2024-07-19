package quiz

import (
	"context"
	"errors"
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
	publicRouter.GET("quizzes/underlying", getQuizzesUnderlyingForUser(service, questionService, engagementService))
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

		// Call InitializeQuizHelper to initialize the quiz
		quizID, err := service.InitializeQuizHelper(
			c.Request.Context(), // Pass the request context to the helper function
			requestData.QuestionIDList,
			requestData.Type,
			requestData.Name,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"quizID": quizID})
	}
}

func (s *QuizService) InitializeQuizHelper(c context.Context, questionIDList []string, quizType *string, quizName *string) (primitive.ObjectID, error) {
	userID, exists := c.Value("userID").(string)
	if !exists {
		return primitive.NilObjectID, errors.New("user ID not found in context")
	}

	// Convert userID to *primitive.ObjectID
	userIDObj, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("invalid user ID: %v", err)
	}

	// Convert questionIDs to ObjectIDs
	questionIDsObjIDs := make([]primitive.ObjectID, len(questionIDList))
	for i, id := range questionIDList {
		questionIDObjID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return primitive.NilObjectID, fmt.Errorf("invalid question ID: %v", err)
		}
		questionIDsObjIDs[i] = questionIDObjID
	}

	if quizType == nil {
		quizType = new(string)
	}

	if quizName == nil {
		name := time.Now().Format("2006-01-02 15:04:05")
		quizName = &name
	}

	quizID, err := s.InitializeQuiz(c, questionIDsObjIDs, userIDObj, quizType, quizName)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("failed to initialize quiz: %v", err)
	}

	return quizID, nil
}

func updateQuiz(service *QuizService) gin.HandlerFunc {
	// request body should contain the quizID and an engagement ID
	return func(c *gin.Context) {

		questionID, err := primitive.ObjectIDFromHex(c.Param("questionID"))
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

		quizID, err = service.UpdateQuiz(c, &quizID, &questionID, &engagementID)

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

		quiz, err := service.GetQuiz(c, quizID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		result, err := service.GetQuizUnderlying(c, service, questionService, engagementService, *quiz)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func getQuizzesUnderlyingForUser(service *QuizService, questionService *question.QuestionService, engagementService *engagement.EngagementService) gin.HandlerFunc {
	return func(c *gin.Context) {
		results, err := GetQuizzesUnderlyingForUser(c, service, questionService, engagementService)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, results)
	}
}

func GetQuizzesUnderlyingForUser(ctx context.Context, service *QuizService, questionService *question.QuestionService, engagementService *engagement.EngagementService) ([]*QuizResult, error) {
	userID, exists := ctx.Value("userID").(string)

	fmt.Println("userID: ", userID)

	if !exists {
		return nil, errors.New("user ID not found in context")
	}

	// Convert userID to *primitive.ObjectID
	userIDObj, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %v", err)
	}

	var quizType *string = nil

	quizzes, err := service.GetQuizzesForUser(ctx, userIDObj, quizType)
	if err != nil {
		return nil, err
	}

	fmt.Println("quizzes length: ", len(quizzes))

	// Print the actual quizzes to inspect their contents
	for i, quiz := range quizzes {
		fmt.Printf("Quiz %d: %+v\n", i, quiz)
	}

	results := make([]*QuizResult, len(quizzes))
	for i, quiz := range quizzes {
		result, err := service.GetQuizUnderlying(ctx, service, questionService, engagementService, *quiz)
		if err != nil {
			return nil, err
		}
		fmt.Println("quiz result i: ", i)
		results[i] = result
	}

	return results, nil
}

func (s *QuizService) GetQuizUnderlying(ctx context.Context, service *QuizService, questionService *question.QuestionService, engagementService *engagement.EngagementService, quiz Quiz) (*QuizResult, error) {

	questionEngagementCombos := make([]QuestionEngagementCombo, len(quiz.QuestionEngagementIDCombos))

	// for each question engagement combo in the quiz, get the question and engagement
	for i, qeid := range quiz.QuestionEngagementIDCombos {
		question, err := questionService.GetQuestion(ctx, *qeid.QuestionID)
		if err != nil {
			return nil, fmt.Errorf("error getting question: %w", err)
		}

		if qeid.EngagementID != nil {
			engagement, err := engagementService.GetEngagementByID(ctx, *qeid.EngagementID)
			if err != nil {
				return nil, fmt.Errorf("error getting engagement: %w", err)
			}

			questionEngagementCombos[i] = QuestionEngagementCombo{
				Question:   question,
				Engagement: engagement,
			}
		} else {
			questionEngagementCombos[i] = QuestionEngagementCombo{
				Question:   question,
				Engagement: nil,
			}
		}
	}

	numTotal := len(quiz.QuestionEngagementIDCombos)
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
		Quiz:            &quiz,
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

			var quizType *string = nil

			quizzes, err := service.GetQuizzesForUser(c, *userIDObj, quizType)

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
