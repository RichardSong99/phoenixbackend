package test

import (
	"example/goserver/engagement"
	"example/goserver/parameterdata"
	"example/goserver/question"
	"example/goserver/quiz"
	"strconv"

	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func RegisterRoutes(publicRouter *gin.RouterGroup, service *TestService, quizService *quiz.QuizService, questionService *question.QuestionService, engagementService *engagement.EngagementService) {
	publicRouter.GET("/test", getTestByName(service))
	publicRouter.GET("/test/:id", getTestByID(service))
	publicRouter.POST("/test", createTest(service))
	publicRouter.GET("/createalltests", createAllTests(service, quizService))
	publicRouter.PATCH("test/:id", updateTest(service))
	publicRouter.GET("/tests", getTestsForUser(service))
	publicRouter.GET("/test/:id/underlying", getTestUnderlying(service, quizService, questionService, engagementService))
	publicRouter.GET("/tests/underlying", getTestsUnderlyingForUser(service, quizService, questionService, engagementService))

}

func getTestByName(service *TestService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusOK, gin.H{"message": "User not logged in"})
			return
		}

		var userIDObj primitive.ObjectID
		var err error
		if userID != nil {
			userIDObj, err = primitive.ObjectIDFromHex(userID.(string))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
				return
			}
		}

		name := c.Query("name")
		test, err := service.GetTestByName(c, name, userIDObj)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, test)
	}
}

func getTestByID(service *TestService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}
		test, err := service.GetTestByID(c, objID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, test)
	}
}

func createTest(service *TestService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var requestData struct {
			Name       string   `json:"Name"`
			QuizIDList []string `json:"QuizIDList"`
		}

		err := c.BindJSON(&requestData)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusOK, gin.H{"message": "User not logged in"})
			return
		}

		var userIDObj primitive.ObjectID

		if userID != nil {
			userIDObj, err = primitive.ObjectIDFromHex(userID.(string))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
				return
			}
		}

		quizIDListObjIDs := make([]primitive.ObjectID, len(requestData.QuizIDList))
		for i, id := range requestData.QuizIDList {
			objID, err := primitive.ObjectIDFromHex(id)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
				return
			}
			quizIDListObjIDs[i] = objID
		}

		// for _, id := range quizIDListObjIDs {
		// 	fmt.Println("quiz ID", id)
		// }

		insertResult, err := service.CreateTest(c, quizIDListObjIDs, requestData.Name, userIDObj)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, insertResult)
	}
}

func getTestsForUser(service *TestService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusOK, gin.H{"message": "User not logged in"})
			return
		}

		var userIDObj primitive.ObjectID
		var err error
		if userID != nil {
			userIDObj, err = primitive.ObjectIDFromHex(userID.(string))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
				return
			}
		}

		tests, err := service.GetTestsForUser(c, userIDObj)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, tests)
	}
}

func getTestsUnderlyingForUser(service *TestService, quizService *quiz.QuizService, questionService *question.QuestionService, engagementService *engagement.EngagementService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusOK, gin.H{"message": "User not logged in"})
			return
		}

		var userIDObj primitive.ObjectID
		var err error
		if userID != nil {
			userIDObj, err = primitive.ObjectIDFromHex(userID.(string))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
				return
			}
		}

		tests, err := service.GetTestsForUser(c, userIDObj)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		testResults := make([]TestResult, len(tests))
		for i, test := range tests {
			testResult, err := service.GetTestUnderlying(c, quizService, questionService, engagementService, test)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			testResults[i] = *testResult
		}

		c.JSON(http.StatusOK, testResults)
	}
}

func updateTest(service *TestService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}

		var requestData struct {
			Completed bool `json:"Completed"`
		}

		err = c.BindJSON(&requestData)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = service.UpdateTest(c, objID, requestData.Completed)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Test updated successfully"})
	}
}

func getTestUnderlying(service *TestService, quizService *quiz.QuizService, questionService *question.QuestionService, engagementService *engagement.EngagementService) gin.HandlerFunc {
	return func(c *gin.Context) {
		testID, err := primitive.ObjectIDFromHex(c.Param("id"))

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid test ID"})
			return
		}

		test, err := service.GetTestByID(c, testID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		testResult, err := service.GetTestUnderlying(c, quizService, questionService, engagementService, *test)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, testResult)

		// c.JSON(http.StatusOK, gin.H{"test": test, "quizResults": quizResults, "testStats": testStats, "mathScaled": mathScaled, "readingScaled": readingScaled, "totalScaled": totalScaled})
	}
}

func (s *TestService) GetTestUnderlying(c *gin.Context, quizService *quiz.QuizService, questionService *question.QuestionService, engagementService *engagement.EngagementService, test Test) (*TestResult, error) {

	quizIDList := test.QuizIDList
	//create array of quiz.QuizResult objects
	quizResults := make([]quiz.QuizResult, len(*quizIDList))
	for i, quizID := range *quizIDList {
		quiz, err := quizService.GetQuiz(c, quizID)
		if err != nil {
			return nil, err
		}

		quizResult, err := quizService.GetQuizUnderlying(c, quizService, questionService, engagementService, *quiz)
		if err != nil {
			return nil, err
		}
		quizResults[i] = *quizResult
	}

	// numMathTotal := 0
	// numReadingTotal := 0
	// numTotalTotal := 0

	// numMathCorrect := 0
	// numReadingCorrect := 0
	// numTotalCorrect := 0

	// type TestStats struct {
	// 	Stats []SmallStats
	// }

	algebraStat := SmallStats{Name: "Algebra", Total: 0, Correct: 0}
	advancedMathStat := SmallStats{Name: "Advanced math", Total: 0, Correct: 0}
	problemSolvingStat := SmallStats{Name: "Problem solving and data analysis", Total: 0, Correct: 0}
	geometryStat := SmallStats{Name: "Geometry and trigonometry", Total: 0, Correct: 0}

	infoIdeasStat := SmallStats{Name: "Information and ideas", Total: 0, Correct: 0}
	craftStructureStat := SmallStats{Name: "Craft and structure", Total: 0, Correct: 0}
	expressionIdeasStat := SmallStats{Name: "Expression of ideas", Total: 0, Correct: 0}
	standardEnglishStat := SmallStats{Name: "Standard English conventions", Total: 0, Correct: 0}

	mathStat := SmallStats{Name: "Math", Total: 0, Correct: 0}
	readingStat := SmallStats{Name: "Reading", Total: 0, Correct: 0}
	totalStat := SmallStats{Name: "Total", Total: 0, Correct: 0}

	for _, topic := range parameterdata.MathTopicsList {
		for _, subtopic := range topic.Children {
			for _, quizResult := range quizResults {
				for _, questionEngCombo := range quizResult.Questions {
					if *questionEngCombo.Question.Topic == subtopic.Name {

						fmt.Println("questionEngCombo", questionEngCombo.Question.Topic, subtopic.Name)

						switch topic.Name {
						case "Algebra":
							algebraStat.Total++
							if questionEngCombo.Engagement != nil && *questionEngCombo.Engagement.Status == "correct" {
								algebraStat.Correct++
							}
						case "Advanced math":
							advancedMathStat.Total++
							if questionEngCombo.Engagement != nil && *questionEngCombo.Engagement.Status == "correct" {
								advancedMathStat.Correct++
							}
						case "Problem solving and data analysis":
							problemSolvingStat.Total++
							if questionEngCombo.Engagement != nil && *questionEngCombo.Engagement.Status == "correct" {
								problemSolvingStat.Correct++
							}
						case "Geometry and trigonometry":
							geometryStat.Total++
							if questionEngCombo.Engagement != nil && *questionEngCombo.Engagement.Status == "correct" {
								geometryStat.Correct++
							}
						}
					}
				}
			}
		}
	}

	// mathStat should be the sum of the other math stats
	mathStat.Total = algebraStat.Total + advancedMathStat.Total + problemSolvingStat.Total + geometryStat.Total
	mathStat.Correct = algebraStat.Correct + advancedMathStat.Correct + problemSolvingStat.Correct + geometryStat.Correct

	for _, topic := range parameterdata.ReadingTopicsList {
		for _, subtopic := range topic.Children {
			for _, quizResult := range quizResults {
				for _, questionEngCombo := range quizResult.Questions {
					if *questionEngCombo.Question.Topic == subtopic.Name {
						switch topic.Name {
						case "Information and ideas":
							infoIdeasStat.Total++
							if questionEngCombo.Engagement != nil && *questionEngCombo.Engagement.Status == "correct" {
								infoIdeasStat.Correct++
							}
						case "Craft and structure":
							craftStructureStat.Total++
							if questionEngCombo.Engagement != nil && *questionEngCombo.Engagement.Status == "correct" {
								craftStructureStat.Correct++
							}
						case "Expression of ideas":
							expressionIdeasStat.Total++
							if questionEngCombo.Engagement != nil && *questionEngCombo.Engagement.Status == "correct" {
								expressionIdeasStat.Correct++
							}
						case "Standard English conventions":
							standardEnglishStat.Total++
							if questionEngCombo.Engagement != nil && *questionEngCombo.Engagement.Status == "correct" {
								standardEnglishStat.Correct++
							}
						}
					}
				}
			}
		}
	}

	// readingStat should be the sum of the other reading stats
	readingStat.Total = infoIdeasStat.Total + craftStructureStat.Total + expressionIdeasStat.Total + standardEnglishStat.Total
	readingStat.Correct = infoIdeasStat.Correct + craftStructureStat.Correct + expressionIdeasStat.Correct + standardEnglishStat.Correct

	// totalStat should be the sum of the math and reading stats
	totalStat.Total = mathStat.Total + readingStat.Total
	totalStat.Correct = mathStat.Correct + readingStat.Correct

	testStats := TestStats{
		Stats: []SmallStats{
			algebraStat,
			advancedMathStat,
			problemSolvingStat,
			geometryStat,
			infoIdeasStat,
			craftStructureStat,
			expressionIdeasStat,
			standardEnglishStat,
			mathStat,
			readingStat,
			totalStat,
		},
	}

	mathScaled := 380
	readingScaled := 380
	totalScaled := mathScaled + readingScaled

	return &TestResult{
		Test:          &test,
		QuizResults:   quizResults,
		TestStats:     &testStats,
		MathScaled:    float64(mathScaled),
		ReadingScaled: float64(readingScaled),
		TotalScaled:   float64(totalScaled),
	}, nil
}

func createAllTests(service *TestService, quizService *quiz.QuizService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusOK, gin.H{"message": "User not logged in"})
			return
		}

		userIDObj, err := primitive.ObjectIDFromHex(userID.(string))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
			return
		}

		// iterate through Tests from parameterdata and create them
		// Iterate through Tests from parameterdata and create them
		for _, test := range parameterdata.Tests {
			_, err := service.CreateTestFromRepresentation(c, *test, userIDObj, quizService)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"message": "All tests created successfully"})
	}
}

func (s *TestService) CreateTestFromRepresentation(c *gin.Context, testRepresentation parameterdata.TestRepresentation, userIDObj primitive.ObjectID, quizService *quiz.QuizService) (primitive.ObjectID, error) {
	quizIDListObjIDs := make([]primitive.ObjectID, len(testRepresentation.QuestionLists))
	for i, questionList := range testRepresentation.QuestionLists {
		quizName := testRepresentation.Name + " - Module " + strconv.Itoa(i+1)
		quizType := "test"
		quizID, err := quizService.InitializeQuizHelper(c, questionList, &quizType, &quizName)
		if err != nil {
			return primitive.NilObjectID, err
		}

		quizIDListObjIDs[i] = quizID
	}

	testID, err := s.CreateTest(c, quizIDListObjIDs, testRepresentation.Name, userIDObj)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return testID, nil
}
