package user

import (
	// ... other imports ...
	"context"
	"errors"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	collection *mongo.Collection
}

func NewUserService(client *mongo.Client) *UserService {
	collection := client.Database("test").Collection("users")
	return &UserService{collection: collection}
}

func (us *UserService) CreateUser(ctx context.Context, user *User) (*User, error) {
	_, err := us.collection.InsertOne(ctx, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (us *UserService) GetUser(ctx context.Context, id string) (*User, error) {
	var user User
	err := us.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (us *UserService) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := us.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (us *UserService) ValidateCredentials(ctx context.Context, email, password string) (*User, error) {
	user, err := us.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid password")
	}

	return user, nil
}

// TokenLifeTime is the duration that a token is valid.
const TokenLifeTime = time.Hour * 24

func (us *UserService) GenerateToken(user *User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"_id": user.ID.Hex(),
		"exp": time.Now().Add(TokenLifeTime).Unix(),
	})

	secretKey := os.Getenv("SECRET_KEY")

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (us *UserService) FetchUserTierFromDB(ctx context.Context, userID string) (string, error) {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return "", err
	}

	var user User
	err = us.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		return "", err
	}

	return user.Tier, nil
}

func (us *UserService) UserExists(c *gin.Context, userID string) (bool, error) {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return false, err
	}

	result := us.collection.FindOne(c, bson.M{"_id": objID})
	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, result.Err()
	}

	return true, nil
}
