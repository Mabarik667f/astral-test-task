package entity

import "github.com/Mabarik667f/fsserver/internal/model"

func ToUser(user User) model.User {
	return model.User{
		ID:    user.ID,
		Login: user.Login,
	}
}

func ToUserWithPasswordHash(user UserWithPasswordHash) model.User {
	return model.User{
		ID:       user.ID,
		Login:    user.Login,
		Password: model.PasswordHash(user.PasswordHash),
	}
}
