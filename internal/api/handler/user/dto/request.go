package dto

type RegisterRequest struct {
	AdminToken string `json:"admin_token" validate:"required"`
	Login      string `json:"login"       validate:"required"`
	Password   string `json:"pswd"        validate:"required"`
}

type LoginRequest struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"pswd"  validate:"required"`
}
