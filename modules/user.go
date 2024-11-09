// File: modules/user.go
package modules

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

// User represents a user in the system
type User struct {
	Username        string
	PasswordHash    string
	Interests       []string
	EnglishLevel    string
	RegistrationDate time.Time
	// Add other relevant fields if necessary
}

// UserManager manages user registrations and authentications
type UserManager struct {
	Users     map[string]User
	JWTSecret string
}

// NewUserManager initializes the UserManager with hardcoded users
func NewUserManager(jwtSecret string) (*UserManager, error) {
	if jwtSecret == "" {
		return nil, errors.New("JWT secret cannot be empty")
	}

	users := make(map[string]User)

	// Hardcoded Users
	hardcodedUsers := []struct {
		Username     string
		Password     string
		Interests    []string
		EnglishLevel string
	}{
		{
			Username:     "user1",
			Password:     "password1",
			Interests:    []string{"math", "science"},
			EnglishLevel: "Intermediate",
		},
		{
			Username:     "user2",
			Password:     "password2",
			Interests:    []string{"history", "literature"},
			EnglishLevel: "Advanced",
		},
	}

	for _, hu := range hardcodedUsers {
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(hu.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("error hashing password for user %s: %v", hu.Username, err)
		}

		users[hu.Username] = User{
			Username:        hu.Username,
			PasswordHash:    string(passwordHash),
			Interests:       hu.Interests,
			EnglishLevel:    hu.EnglishLevel,
			RegistrationDate: time.Now(),
		}
	}

	return &UserManager{
		Users:     users,
		JWTSecret: jwtSecret,
	}, nil
}

// RegisterUser registers a new user (not necessary if hardcoding, but kept for completeness)
func (um *UserManager) RegisterUser(username, password string, interests []string, englishLevel string) error {
	if _, exists := um.Users[username]; exists {
		return errors.New("username already exists")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error hashing password: %v", err)
	}

	um.Users[username] = User{
		Username:        username,
		PasswordHash:    string(passwordHash),
		Interests:       interests,
		EnglishLevel:    englishLevel,
		RegistrationDate: time.Now(),
	}

	return nil
}

// AuthenticateUser verifies the username and password and returns a JWT token
func (um *UserManager) AuthenticateUser(username, password string) (string, error) {
	user, exists := um.Users[username]
	if !exists {
		return "", errors.New("invalid username or password")
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", errors.New("invalid username or password")
	}

	// Create JWT Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
	})

	tokenString, err := token.SignedString([]byte(um.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("error signing token: %v", err)
	}

	return tokenString, nil
}

// ValidateToken validates the JWT token and returns the username
func (um *UserManager) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure the token's signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(um.JWTSecret), nil
	})

	if err != nil {
		return "", fmt.Errorf("invalid token: %v", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		username, ok := claims["username"].(string)
		if !ok {
			return "", errors.New("username not found in token")
		}
		return username, nil
	}

	return "", errors.New("invalid token claims")
}

// GetUserContext retrieves the user context for the given username
func (um *UserManager) GetUserContext(username string) (User, error) {
	user, exists := um.Users[username]
	if !exists {
		return User{}, errors.New("user not found")
	}
	return user, nil
}
