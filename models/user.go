package models

type User struct {
	ID int `json:"id"`
	Email string `json:"email"`
	Password string `json:"password"`
	Role string `json:"role"`
}
type LoginRequest struct {
	Email string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}
type RegisterRequest struct {
	Email string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}