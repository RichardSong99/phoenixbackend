package user

import (
	"net/http"
	"net/mail"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	// Add other fields as needed
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterRoutes registers the user routes.
func RegisterRoutes(router *gin.Engine, userService *UserService) {
	userGroup := router.Group("/user")
	{
		userGroup.POST("/register", func(c *gin.Context) { createUser(c, userService) })
		userGroup.GET("/:id", func(c *gin.Context) { getUser(c, userService) })
		userGroup.POST("/login", func(c *gin.Context) { loginUser(c, userService) })
		userGroup.GET("/confirm", func(c *gin.Context) { confirmUser(c, userService) }) // Removed :id

		// Add more routes as needed
	}
}

// GetUserTier fetches the user's tier from the database
func (us *UserService) GetUserTier(c *gin.Context) string {
	// Attempt to get user ID from JWT token
	userID, ok := c.Get("userID")

	userTier := "free" // Default to free tier
	if ok {
		// Fetch user's tier from the database
		tier, err := us.FetchUserTierFromDB(c.Request.Context(), userID.(string))
		if err == nil {
			userTier = tier
		}
		// If there's an error, it could be due to an invalid userID, so default to free tier
	}

	return userTier
}

// createUser handles the creation of a new user.
func createUser(c *gin.Context, userService *UserService) {
	var request RegisterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if the email is valid
	_, err := mail.ParseAddress(request.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email address"})
		return
	}

	// Check if a user with the same email already exists
	existingUser, err := userService.GetUserByEmail(c, request.Email)
	if err != nil {
		if err.Error() == "user not found" {
			// User not found, continue execution
		} else {
			// Other error, return an internal server error
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check if email is already in use"})
			return
		}
	}
	if existingUser != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is already in use"})
		return
	}

	// Check if the password is at least 8 characters long
	if len(request.Password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 8 characters long"})
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create the user with the hashed password
	newUser := User{
		Email:        request.Email,
		PasswordHash: string(hashedPassword),
		// Add other fields as needed
	}

	createdUser, err := userService.CreateUser(c, &newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdUser)
}

// getUser handles getting a user by their ID.
func getUser(c *gin.Context, userService *UserService) {
	id := c.Param("id")

	user, err := userService.GetUser(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// loginUser handles the login of a user.
func loginUser(c *gin.Context, userService *UserService) {
	var request LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate the user's credentials
	user, err := userService.ValidateCredentials(c, request.Email, request.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate a token for the user
	token, err := userService.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "user_id": user.ID.Hex()})
}

func confirmUser(c *gin.Context, userService *UserService) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID does not exist"})
		return
	}

	exists, err := userService.UserExists(c, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"userExists": exists})
}
