package service

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"todo-api/internal/config"
	"todo-api/internal/errors"
	"todo-api/internal/model"
	"todo-api/internal/repository"
)

// AuthService handles authentication logic
type AuthService struct {
	userRepo     *repository.UserRepository
	denylistRepo *repository.JwtDenylistRepository
	config       *config.Config
}

// NewAuthService creates a new AuthService
func NewAuthService(
	userRepo *repository.UserRepository,
	denylistRepo *repository.JwtDenylistRepository,
	cfg *config.Config,
) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		denylistRepo: denylistRepo,
		config:       cfg,
	}
}

// JWTClaims represents the claims in a JWT token (Rails devise-jwt compatible)
type JWTClaims struct {
	Sub string `json:"sub"` // User ID
	Jti string `json:"jti"` // Token identifier
	Scp string `json:"scp"` // Scope
	jwt.RegisteredClaims
}

// SignUp registers a new user and returns the user and JWT token
func (s *AuthService) SignUp(email, password, passwordConfirmation, name string) (*model.User, string, error) {
	// Validate password confirmation
	if password != passwordConfirmation {
		return nil, "", errors.ValidationFailed(map[string][]string{
			"password_confirmation": {"doesn't match Password"},
		})
	}

	// Check for duplicate email
	exists, err := s.userRepo.ExistsByEmail(email)
	if err != nil {
		return nil, "", errors.InternalError()
	}
	if exists {
		return nil, "", errors.DuplicateResource("User", "email")
	}

	// Create user
	user := &model.User{
		Email: email,
		Name:  &name,
	}
	if err := user.SetPassword(password); err != nil {
		return nil, "", errors.InternalError()
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, "", errors.InternalError()
	}

	// Generate token
	token, err := s.GenerateToken(user)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

// SignIn authenticates a user and returns the user and JWT token
func (s *AuthService) SignIn(email, password string) (*model.User, string, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, "", errors.AuthenticationFailed("Invalid email or password")
	}

	if !user.CheckPassword(password) {
		return nil, "", errors.AuthenticationFailed("Invalid email or password")
	}

	token, err := s.GenerateToken(user)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

// SignOut revokes the given token by adding its jti to the denylist
func (s *AuthService) SignOut(jti string, exp time.Time) error {
	return s.denylistRepo.Add(jti, exp)
}

// GenerateToken creates a new JWT token for the given user
func (s *AuthService) GenerateToken(user *model.User) (string, error) {
	now := time.Now()
	expiration := now.Add(time.Duration(s.config.JWTExpirationHours) * time.Hour)

	claims := JWTClaims{
		Sub: fmt.Sprintf("%d", user.ID),
		Jti: uuid.New().String(),
		Scp: "user",
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiration),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWTSecret))
}

// ValidateToken validates a JWT token and returns the claims
func (s *AuthService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Check if token is revoked
	revoked, err := s.denylistRepo.Exists(claims.Jti)
	if err != nil {
		return nil, err
	}
	if revoked {
		return nil, fmt.Errorf("token has been revoked")
	}

	return claims, nil
}
