package errs

import "errors"

var (
	ErrAdminToken = errors.New("admin token not valid")

	ErrLoginPatternMatch = errors.New("login not matched to pattern")

	ErrPasswordsNotEqual = errors.New("passwords not equal")

	ErrPasswordTooShort  = errors.New("password must be at least 8 characters")
	ErrPasswordNoUpper   = errors.New("password must contain at least one uppercase letter")
	ErrPasswordNoLower   = errors.New("password must contain at least one lowercase letter")
	ErrPasswordNoDigit   = errors.New("password must contain at least one digit")
	ErrPasswordNoSpecial = errors.New("password must contain at least one special character")
	ErrPasswordHasSpaces = errors.New("password must not contain spaces")

	ErrUserUnique   = errors.New("user fields unique constraint")
	ErrUserNotFound = errors.New("user not found")
)
