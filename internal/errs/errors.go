package errs

import "errors"

var (
	ErrInvalidJSON    = errors.New("invalid json body")
	ErrJSONValidation = errors.New("invalid json format")

	ErrInternalServer = errors.New("internal server error")

	ErrAdminToken = errors.New("admin token not valid")

	ErrTokenNotProvided = errors.New("token was not provided")
	ErrSessionNoExists  = errors.New("session does not exists")

	ErrLoginPatternMatch   = errors.New("login not matched to pattern")
	ErrPasswordPatterMatch = errors.New("password not matched to pattern")

	ErrPasswordTooShort  = errors.New("password must be at least 8 characters")
	ErrPasswordNoUpper   = errors.New("password must contain at least one uppercase letter")
	ErrPasswordNoLower   = errors.New("password must contain at least one lowercase letter")
	ErrPasswordNoDigit   = errors.New("password must contain at least one digit")
	ErrPasswordNoSpecial = errors.New("password must contain at least one special character")
	ErrPasswordHasSpaces = errors.New("password must not contain spaces")

	ErrPasswordsNotEqual = errors.New("passwords not equal")

	ErrDocNameEmpty     = errors.New("document name can not be empty")
	ErrDocMimeTypeEmpty = errors.New("document mime type can not be empty")
	ErrDocBusiness      = errors.New("document failed to create: business error")

	ErrSaveFileToStorage = errors.New("error to save file in storage")

	ErrUserUnique   = errors.New("user fields unique constraint")
	ErrUserNotFound = errors.New("user not found")
)
