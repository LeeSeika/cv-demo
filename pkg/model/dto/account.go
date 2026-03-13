package dto

type CreateAccountReq struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UpdateAccountReq struct {
	Name        string `json:"name" binding:"required"`
	AvatarURL   string `json:"avatar_url" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type AccountInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}
