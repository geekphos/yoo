package v1

import "time"

type LoginRequest struct {
	Email    string `json:"email" binding:"email,required"`
	Password string `json:"password" binding:"required,min=6,max=14"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}

type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required,min=6,max=14"`
	NewPassword string `json:"newPassword" binding:"required,min=6,max=14"`
}

type CreateUserRequest struct {
	Email    string `json:"email" binding:"email,required"`
	Nickname string `json:"nickname" binding:"required"`
	Password string `json:"password" binding:"required,min=6,max=14"`
}

type ProfileResponse struct {
	ID        int32     `json:"id"`
	Email     string    `json:"email"`
	Nickname  string    `json:"nickname"`
	Avatar    string    `json:"avatar"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
