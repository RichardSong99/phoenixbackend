package main

import (
	"context"
	"example/goserver/datacube"
	"example/goserver/engagement"
	"example/goserver/lessons"
	"example/goserver/parameterdata"
	"example/goserver/question" // replace with your project path
	"example/goserver/quiz"     // replace with your project path
	"example/goserver/test"
	"example/goserver/upload" // replace with your project path
	"example/goserver/user"   // replace with your project path
	"example/goserver/video"
	"example/goserver/videoengagement"

	youtubePackage "example/goserver/youtubePackage" // replace with your project path
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file in main")
	}

	// Set up MongoDB client
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("ATLAS_URI")))
	if err != nil {
		panic(err)
	}
	defer client.Disconnect(ctx)

	// Create a new UserService
	userService := user.NewUserService(client)

	// Create a new QuestionService
	questionService := question.NewQuestionService(client)

	// Create a new EngagementService
	engagementService := engagement.NewEngagementService(client) // Remove questionService parameter

	// Create a new DataCubeService
	dataCubeService := datacube.NewDataCubeService(client, questionService)

	lessonService := lessons.NewLessonService(client)
	courseService := lessons.NewCourseService(client)

	var youtubeService *youtubePackage.YouTubeService

	for i := 0; i < 5; i++ {
		youtubeService, err = youtubePackage.NewYouTubeService(ctx)
		if err != nil {
			fmt.Println("Error creating YouTube service:", err)
			time.Sleep(1 * time.Second) // wait for 1 second before retrying
		} else {
			break
		}
	}

	if err != nil {
		fmt.Println("Failed to create YouTube service after 5 attempts:", err)
		return
	}

	videoService := video.NewVideoService(client)

	videoEngagementService := videoengagement.NewVideoEngagementService(client)

	quizService, err := quiz.NewQuizService(ctx, client)
	if err != nil {
		fmt.Println("Error creating quiz service:", err)
		return
	}

	testService, err := test.NewTestService(ctx, client)
	if err != nil {
		fmt.Println("Error creating test service:", err)
		return
	}

	// Set up Gin router
	router := gin.Default()

	// Apply the middleware to the router (insert this line)
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// Add the JWT middleware
	router.Use(user.JWTMiddleware(userService))

	// Group for public (unauthenticated) routes
	publicRoutes := router.Group("/")

	// Register routes
	user.RegisterRoutes(router, userService)
	engagement.RegisterRoutes(router, engagementService)

	// Add the upload route
	publicRoutes.POST("/upload", func(c *gin.Context) {
		upload.UploadHandler(c.Writer, c.Request)
	})

	// Create a group for routes that require JWT authentication
	authenticated := router.Group("/")
	authenticated.Use(user.JWTMiddleware(userService))
	// Register routes that require authentication
	question.RegisterRoutes(publicRoutes, authenticated, questionService, userService)

	lessons.RegisterRoutes(publicRoutes, lessonService, courseService)

	// Add the datacube routes
	datacube.RegisterRoutes(publicRoutes, authenticated, dataCubeService)

	parameterdata.RegisterRoutes(publicRoutes, nil)

	// Move routes that require authentication to the authenticated group
	// For example, if you have a route for getting a user's profile that requires authentication:
	// authenticated.GET("/user/:id", func(c *gin.Context) { getUser(c, userService) })

	video.RegisterRoutes(publicRoutes, videoService, youtubeService.Service)

	videoengagement.RegisterRoutes(publicRoutes, videoEngagementService)

	quiz.RegisterRoutes(publicRoutes, quizService, questionService, engagementService)

	test.RegisterRoutes(publicRoutes, testService, quizService, questionService, engagementService)

	// Start server
	router.Run(":8080")
}
