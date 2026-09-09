package dto

type RegisterResponse struct {
	Login string `json:"login"`
}

type LoginResponse struct {
	Token string `json:"token"`
}
