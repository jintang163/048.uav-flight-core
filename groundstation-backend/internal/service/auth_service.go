package service

import (
	"context"
	"errors"
	"groundstation-backend/internal/config"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/repository"
	"groundstation-backend/pkg/utils"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	userRepo *repository.UserRepository
}

func NewAuthService() *AuthService {
	return &AuthService{
		userRepo: repository.NewUserRepository(),
	}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone"`
	FullName string `json:"full_name"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	User         *models.User `json:"user"`
}

type JWTClaims struct {
	UserID       uint64           `json:"user_id"`
	Username     string           `json:"username"`
	Role         models.UserRole  `json:"role"`
	TokenVersion int              `json:"token_version"`
	jwt.RegisteredClaims
}

func (s *AuthService) Login(req *LoginRequest, ip, userAgent string) (*TokenResponse, error) {
	user, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	if user.Status == models.UserStatusLocked {
		return nil, errors.New("account is locked")
	}

	if user.Status == models.UserStatusInactive {
		return nil, errors.New("account is inactive")
	}

	if !utils.CheckPassword(req.Password, user.PasswordSalt, user.PasswordHash) {
		_ = s.userRepo.IncrementLoginAttempts(req.Username)
		if user.LoginAttempts >= 5 {
			_ = s.userRepo.LockUser(user.ID)
		}
		return nil, errors.New("invalid username or password")
	}

	_ = s.userRepo.UpdateLoginInfo(user.ID, ip)

	loginLog := &models.LoginLog{
		UserID:    user.ID,
		Username:  user.Username,
		Success:   true,
		IPAddress: ip,
		UserAgent: userAgent,
	}
	_ = s.userRepo.CreateLoginLog(loginLog)

	return s.generateTokens(user)
}

func (s *AuthService) Register(req *RegisterRequest) (*models.User, error) {
	exists, _ := s.userRepo.CheckUsernameExists(req.Username, 0)
	if exists {
		return nil, errors.New("username already exists")
	}

	if req.Email != "" {
		exists, _ = s.userRepo.CheckEmailExists(req.Email, 0)
		if exists {
			return nil, errors.New("email already exists")
		}
	}

	user := &models.User{
		Username:     req.Username,
		PasswordHash: req.Password,
		Email:        req.Email,
		Phone:        req.Phone,
		FullName:     req.FullName,
		Role:         models.RoleUser,
		Status:       models.UserStatusActive,
		TokenVersion: 1,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	user.PasswordHash = ""
	user.PasswordSalt = ""
	return user, nil
}

func (s *AuthService) generateTokens(user *models.User) (*TokenResponse, error) {
	accessExpire := time.Duration(config.AppConfig.JWT.AccessTokenExpire) * time.Second
	refreshExpire := time.Duration(config.AppConfig.JWT.RefreshTokenExpire) * time.Second

	accessClaims := &JWTClaims{
		UserID:       user.ID,
		Username:     user.Username,
		Role:         user.Role,
		TokenVersion: user.TokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessExpire)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "groundstation",
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenStr, err := accessToken.SignedString([]byte(config.AppConfig.JWT.Secret))
	if err != nil {
		return nil, err
	}

	refreshClaims := &JWTClaims{
		UserID:       user.ID,
		Username:     user.Username,
		Role:         user.Role,
		TokenVersion: user.TokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(refreshExpire)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "groundstation",
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenStr, err := refreshToken.SignedString([]byte(config.AppConfig.JWT.Secret))
	if err != nil {
		return nil, err
	}

	_ = s.userRepo.UpdateRefreshToken(user.ID, refreshTokenStr)

	return &TokenResponse{
		AccessToken:  accessTokenStr,
		RefreshToken: refreshTokenStr,
		TokenType:    "Bearer",
		ExpiresIn:    config.AppConfig.JWT.AccessTokenExpire,
		User:         user,
	}, nil
}

func (s *AuthService) RefreshToken(refreshToken string) (*TokenResponse, error) {
	claims := &JWTClaims{}
	token, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.AppConfig.JWT.Secret), nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid refresh token")
	}

	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if user.TokenVersion != claims.TokenVersion {
		return nil, errors.New("token has been revoked")
	}

	if user.RefreshToken != refreshToken {
		return nil, errors.New("refresh token not match")
	}

	return s.generateTokens(user)
}

func (s *AuthService) ValidateToken(tokenString string) (*JWTClaims, error) {
	claims := &JWTClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.AppConfig.JWT.Secret), nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	if claims.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("token expired")
	}

	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if user.TokenVersion != claims.TokenVersion {
		return nil, errors.New("token has been revoked")
	}

	if user.Status != models.UserStatusActive {
		return nil, errors.New("user account is not active")
	}

	return claims, nil
}

func (s *AuthService) Logout(userID uint64) error {
	return s.userRepo.IncrementTokenVersion(userID)
}

func (s *AuthService) ChangePassword(userID uint64, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	if !utils.CheckPassword(oldPassword, user.PasswordSalt, user.PasswordHash) {
		return errors.New("old password is incorrect")
	}

	return s.userRepo.UpdatePassword(userID, newPassword)
}

func (s *AuthService) GetUserByID(userID uint64) (*models.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	user.PasswordHash = ""
	user.PasswordSalt = ""
	user.RefreshToken = ""
	return user, nil
}

func (s *AuthService) CheckPermission(role models.UserRole, allowedRoles []models.UserRole) bool {
	for _, r := range allowedRoles {
		if r == role {
			return true
		}
	}
	return false
}

func (s *AuthService) GetUserFromContext(ctx context.Context) *JWTClaims {
	claims, ok := ctx.Value("user").(*JWTClaims)
	if !ok {
		return nil
	}
	return claims
}
