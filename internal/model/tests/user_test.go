package tests

import (
	"testing"

	"github.com/Mabarik667f/fsserver/internal/errs"
	"github.com/Mabarik667f/fsserver/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewUser(t *testing.T) {
	type args struct {
		passwordHash model.PasswordHash
		login        string
	}

	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "success",
			args: args{
				passwordHash: model.PasswordHash("Valid+Hash123"),
				login:        "Validlogin1",
			},
			wantErr: nil,
		},
		{
			name: "invalid login",
			args: args{
				passwordHash: model.PasswordHash("Valid+Hash123"),
				login:        "invalidю",
			},
			wantErr: errs.ErrLoginPatternMatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			user, err := model.NewUser(tt.args.passwordHash, tt.args.login)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, user)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, user)

			assert.NotEqual(t, uuid.Nil, user.ID)
			assert.Equal(t, tt.args.login, user.Login)
			assert.Equal(t, tt.args.passwordHash, user.Password)
		})
	}
}

func Test_ValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{
			name:     "success",
			password: "Valid+Hash123",
			wantErr:  nil,
		},
		{
			name:     "too short",
			password: "len",
			wantErr:  errs.ErrPasswordTooShort,
		},
		{
			name:     "no upper",
			password: "noupperpassword+123",
			wantErr:  errs.ErrPasswordNoUpper,
		},
		{
			name:     "no lower",
			password: "NOLOWERPASSWORD+123",
			wantErr:  errs.ErrPasswordNoLower,
		},
		{
			name:     "no digit",
			password: "noDigitPassword+",
			wantErr:  errs.ErrPasswordNoDigit,
		},
		{
			name:     "no special",
			password: "noDigitPassword123",
			wantErr:  errs.ErrPasswordNoSpecial,
		},
		{
			name:     "has space",
			password: "PasswordWithS+pace1234 s",
			wantErr:  errs.ErrPasswordHasSpaces,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := model.ValidatePassword(tt.password)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)

				return
			}

			require.NoError(t, err)
		})
	}
}
