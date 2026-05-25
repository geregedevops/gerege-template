// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"templatev27/internal/apperror"
	"templatev27/internal/business/usecases/auth"
	"templatev27/internal/business/usecases/users"
	"templatev27/pkg/verify"
)

func TestVerifyOTP(t *testing.T) {
	tests := []struct {
		name  string
		email string
		code  string
		setup func(f *fixture)
		// wantErr / wantErrType-ийг хослуулсан, учир нь apperror.ErrTypeInternal
		// нь iota-гийн тэг — ганц sentinel нь тэр төрөлтэй мөргөлдөж, чимээгүйхэн
		// тэнцэх байсан.
		wantErr     bool
		wantErrType apperror.ErrorType
	}{
		{
			name:  "happy path consumes request_id, activates user, clears attempt key",
			email: "patrick@example.com",
			code:  "123456",
			setup: func(f *fixture) {
				user := activeUser(t)
				user.Active = false
				f.users.On("GetByEmail", mock.Anything, users.GetByEmailRequest{Email: "patrick@example.com"}).Return(users.GetByEmailResponse{User: user}, nil).Once()
				f.redis.On("Incr", mock.Anything, "otp_attempts:patrick@example.com").Return(int64(1), nil).Once()
				f.redis.On("Expire", mock.Anything, "otp_attempts:patrick@example.com", mock.AnythingOfType("time.Duration")).Return(nil).Once()
				// request_id-г GetDel-ээр атомар уншиж устгана.
				f.redis.On("GetDel", mock.Anything, "user_otp:patrick@example.com").Return("clv_abc123", nil).Once()
				f.verify.On("Check", mock.Anything, "clv_abc123", "123456").Return(nil).Once()
				f.users.On("Activate", mock.Anything, users.ActivateRequest{UserID: user.ID}).Return(nil).Once()
				f.redis.On("Del", mock.Anything, "otp_attempts:patrick@example.com").Return(nil).Once()
			},
		},
		{
			name:  "wrong code returns BadRequest, request_id rehydrated for further attempts",
			email: "patrick@example.com",
			code:  "999999",
			setup: func(f *fixture) {
				user := activeUser(t)
				user.Active = false
				f.users.On("GetByEmail", mock.Anything, users.GetByEmailRequest{Email: "patrick@example.com"}).Return(users.GetByEmailResponse{User: user}, nil).Once()
				f.redis.On("Incr", mock.Anything, "otp_attempts:patrick@example.com").Return(int64(1), nil).Once()
				f.redis.On("Expire", mock.Anything, "otp_attempts:patrick@example.com", mock.AnythingOfType("time.Duration")).Return(nil).Once()
				f.redis.On("GetDel", mock.Anything, "user_otp:patrick@example.com").Return("clv_abc123", nil).Once()
				f.verify.On("Check", mock.Anything, "clv_abc123", "999999").Return(verify.ErrNotApproved).Once()
				f.redis.On("SetWithTTL", mock.Anything, "user_otp:patrick@example.com", "clv_abc123", mock.AnythingOfType("time.Duration")).Return(nil).Once()
			},
			wantErr:     true,
			wantErrType: apperror.ErrTypeBadRequest,
		},
		{
			name:  "missing request_id surfaces as BadRequest expired",
			email: "patrick@example.com",
			code:  "123456",
			setup: func(f *fixture) {
				user := activeUser(t)
				user.Active = false
				f.users.On("GetByEmail", mock.Anything, users.GetByEmailRequest{Email: "patrick@example.com"}).Return(users.GetByEmailResponse{User: user}, nil).Once()
				f.redis.On("Incr", mock.Anything, "otp_attempts:patrick@example.com").Return(int64(1), nil).Once()
				f.redis.On("Expire", mock.Anything, "otp_attempts:patrick@example.com", mock.AnythingOfType("time.Duration")).Return(nil).Once()
				f.redis.On("GetDel", mock.Anything, "user_otp:patrick@example.com").Return("", errors.New("redis: nil")).Once()
			},
			wantErr:     true,
			wantErrType: apperror.ErrTypeBadRequest,
		},
		{
			name:  "lockout after exceeding OTPMaxAttempts surfaces as Forbidden",
			email: "patrick@example.com",
			code:  "123456",
			setup: func(f *fixture) {
				user := activeUser(t)
				user.Active = false
				f.users.On("GetByEmail", mock.Anything, users.GetByEmailRequest{Email: "patrick@example.com"}).Return(users.GetByEmailResponse{User: user}, nil).Once()
				// 6 дахь оролдлого OTPMaxAttempts=5-аас давна; Incr нь 6
				// буцаана. Forbidden нь BadRequest "буруу код"-оос ялгаатай
				// дохио тул rate-limit / сэрэмжлүүлэг нь brute-force загварыг
				// бичгийн алдаанаас ялгаж чадна.
				f.redis.On("Incr", mock.Anything, "otp_attempts:patrick@example.com").Return(int64(6), nil).Once()
				// attempts != 1 тул incrWithExpiry нь TTL байгаа эсэхийг
				// PTTL-ээр шалгана; эерэг утга буцаавал дахин Expire хийхгүй.
				f.redis.On("PTTL", mock.Anything, "otp_attempts:patrick@example.com").Return(5*time.Minute, nil).Once()
			},
			wantErr:     true,
			wantErrType: apperror.ErrTypeForbidden,
		},
		{
			name:  "already-active account short-circuits with BadRequest",
			email: "patrick@example.com",
			code:  "123456",
			setup: func(f *fixture) {
				// Аль хэдийн идэвхтэй — эрт буцалт, Incr / Get / Activate байхгүй.
				f.users.On("GetByEmail", mock.Anything, users.GetByEmailRequest{Email: "patrick@example.com"}).Return(users.GetByEmailResponse{User: activeUser(t)}, nil).Once()
			},
			wantErr:     true,
			wantErrType: apperror.ErrTypeBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t)
			tt.setup(f)

			err := f.usecase.VerifyOTP(context.Background(), auth.VerifyOTPRequest{Email: tt.email, OTPCode: tt.code})

			if !tt.wantErr {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			var domErr *apperror.DomainError
			require.True(t, errors.As(err, &domErr))
			assert.Equal(t, tt.wantErrType, domErr.Type)
		})
	}
}
