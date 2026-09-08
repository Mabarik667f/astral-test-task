package model

import (
	"regexp"
	"unicode"

	"github.com/Mabarik667f/fsserver/internal/errs"
	"github.com/google/uuid"
)

type PasswordHash string

type User struct {
	ID       uuid.UUID
	Login    string
	Password PasswordHash
}

func NewUser(
	passwordHash PasswordHash, login string,
) (*User, error) {
	if err := validateLogin(login); err != nil {
		return nil, err
	}

	return &User{
		ID:       uuid.New(),
		Login:    login,
		Password: passwordHash,
	}, nil
}

const loginPattern = `^[A-Za-z0-9]{8,}$`

func validateLogin(login string) error {
	matched, err := regexp.MatchString(loginPattern, login)
	if err != nil {
		return err
	}
	if !matched {
		return errs.ErrLoginPatternMatch
	}
	return nil
}

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errs.ErrPasswordTooShort
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasDigit   bool
		hasSpecial bool
	)

	for _, r := range password {
		switch {
		case unicode.IsSpace(r):
			return errs.ErrPasswordHasSpaces
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		default:
			hasSpecial = true
		}
	}

	if !hasUpper {
		return errs.ErrPasswordNoUpper
	}

	if !hasLower {
		return errs.ErrPasswordNoLower
	}

	if !hasDigit {
		return errs.ErrPasswordNoDigit
	}

	if !hasSpecial {
		return errs.ErrPasswordNoSpecial
	}

	return nil
}
