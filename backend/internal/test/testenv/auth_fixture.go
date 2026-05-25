//go:build integration

// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package testenv

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"templatev27/internal/business/usecases/auth"
	"templatev27/internal/business/usecases/users"
	"templatev27/internal/config"
	"templatev27/internal/datasources/caches"
	userspostgres "templatev27/internal/datasources/repositories/postgres/users"
	"templatev27/pkg/helpers"
	"templatev27/pkg/jwt"
	"templatev27/pkg/verify"
)

// AuthFixture нь end-to-end тестүүдэд ашиглагддаг бүрэн холбогдсон auth
// хэсэг юм: жинхэнэ Postgres, жинхэнэ Redis, жинхэнэ Ristretto, жинхэнэ
// JWT — зөвхөн гадагш чиглэсэн SMTP mailer л хуурамчаар хийгдсэн, учир
// нь бид OTP кодуудыг барьж аваад VerifyOTP руу буцаан өгөх хэрэгтэй
// бөгөөд SMTP-г CI-д ажиллуулах нь ямар ч байсан үнэ цэнэгүй.
//
// Хоёр bounded context хоёулаа илчлэгдсэн: туршиж буй auth урсгалуудад
// Auth, хэрэглэгчийн бичлэгүүдийг шууд унших эсвэл өөрчлөх шаардлагатай
// аливаа тохиргоо / баталгаажуулалтын алхамд Users.
type AuthFixture struct {
	Auth     auth.Usecase
	Users    users.Usecase
	Mailer   *CapturingMailer
	Verifier *FakeVerifier
	JWT      jwt.JWTService
}

// FakeVerifier нь verify.Sender-ийг хэрэгжүүлдэг локал бодит хэлбэр —
// integration тестүүд OTP-ийг алсын API-руу хийхгүйгээр код-руу хүрнэ.
// Send нь дотооддоо 6 оронтой код үүсгэж, request_id-той хамт хадгална;
// Check нь тэр кодыг хэрэглэгчийн оруулсантай тулгана. Нэг удаагийн
// semantic-ийг хадгалахын тулд амжилттай Check нь request_id-г устгана.
type FakeVerifier struct {
	mu       sync.Mutex
	issued   map[string]otpCapture
	captured []otpCapture
}

func newFakeVerifier() *FakeVerifier {
	return &FakeVerifier{issued: make(map[string]otpCapture)}
}

// Send нь шинэ код үүсгэж, олгосон request_id буцаана. Алсын verify.gecloud.mn-ийг орлоно.
func (v *FakeVerifier) Send(_ context.Context, destination string) (string, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	code, err := helpers.GenerateOTPCode(6)
	if err != nil {
		return "", err
	}
	rid := "fake_" + uuid.NewString()
	entry := otpCapture{Code: code, Receiver: destination, RequestID: rid}
	v.issued[rid] = entry
	v.captured = append(v.captured, entry)
	return rid, nil
}

// Check нь хэрэглэгчийн оруулсан кодыг request_id-ийн дор хадгалсан кодтой тулгана.
func (v *FakeVerifier) Check(_ context.Context, requestID, code string) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	entry, ok := v.issued[requestID]
	if !ok || entry.Code != code {
		return verify.ErrNotApproved
	}
	delete(v.issued, requestID)
	return nil
}

// LastOTP нь хүлээн авагчид зориулж хамгийн сүүлд "илгээсэн" OTP-г буцаана.
func (v *FakeVerifier) LastOTP(t *testing.T, receiver string) string {
	t.Helper()
	v.mu.Lock()
	defer v.mu.Unlock()
	for i := len(v.captured) - 1; i >= 0; i-- {
		if v.captured[i].Receiver == receiver {
			return v.captured[i].Code
		}
	}
	t.Fatalf("no OTP captured for %s", receiver)
	return ""
}

// CapturingMailer нь OTP+хүлээн авагч хос бүрийг бүртгэдэг тул тестүүд
// 6 санамсаргүй оронг таахын оронд кодыг гаргаж авч чадна.
type CapturingMailer struct {
	mu       sync.Mutex
	captured []otpCapture
}

type otpCapture struct{ Code, Receiver, RequestID string }

// SendOTP нь mailer.OTPMailer-г хангадаг. Үргэлж амжилттай болдог.
func (m *CapturingMailer) SendOTP(_ context.Context, code, receiver string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.captured = append(m.captured, otpCapture{Code: code, Receiver: receiver})
	return nil
}

func (m *CapturingMailer) SendPasswordReset(ctx context.Context, token, receiver string) error {
	return m.SendOTP(ctx, token, receiver)
}

// LastOTP нь хүлээн авагчид зориулж хамгийн сүүлд барьж авсан OTP-г
// буцаана, эсвэл нэг ч илгээгээгүй бол тестийг унагана.
func (m *CapturingMailer) LastOTP(t *testing.T, receiver string) string {
	t.Helper()
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := len(m.captured) - 1; i >= 0; i-- {
		if m.captured[i].Receiver == receiver {
			return m.captured[i].Code
		}
	}
	t.Fatalf("no OTP captured for %s", receiver)
	return ""
}

// NewAuthFixture нь хоёр bounded context-г шинэ Postgres + Redis
// контейнеруудтай холбоно. Тохируулж болох тохиргоонууд (OTP оролдлого,
// JWT secret-ийн урт, bcrypt cost) нь боломжийн өгөгдмөл утгуудаас
// seed хийгддэг — тэдгээрийг өөрчлөх шаардлагатай тестүүд дуудахаасаа
// өмнө config.AppConfig-г шууд дарж бичиж болно.
func NewAuthFixture(t *testing.T) *AuthFixture {
	t.Helper()
	db := StartPostgres(t)
	redis := StartRedis(t)

	if config.AppConfig.OTPMaxAttempts == 0 {
		config.AppConfig.OTPMaxAttempts = 5
	}
	if config.AppConfig.REDISExpired == 0 {
		config.AppConfig.REDISExpired = 5
	}
	if config.AppConfig.BcryptCost == 0 {
		// register нь дуудалт бүрт 100ms+ нэмэхгүй байхын тулд тестүүдэд
		// cost-г бууруул.
		config.AppConfig.BcryptCost = 4
	}
	if config.AppConfig.JWTSecret == "" {
		config.AppConfig.JWTSecret = "integration-test-secret-thirty-two-chars!"
	}
	if config.AppConfig.JWTIssuer == "" {
		config.AppConfig.JWTIssuer = "integration-test"
	}
	if config.AppConfig.JWTExpired == 0 {
		config.AppConfig.JWTExpired = 1
	}
	if config.AppConfig.JWTRefreshExpired == 0 {
		config.AppConfig.JWTRefreshExpired = 7
	}

	ristretto, err := caches.NewRistrettoCache()
	require.NoError(t, err)

	jwtSvc := jwt.NewJWTServiceWithRefresh(
		config.AppConfig.JWTSecret,
		config.AppConfig.JWTIssuer,
		config.AppConfig.JWTExpired,
		config.AppConfig.JWTRefreshExpired,
	)

	mailer := &CapturingMailer{}
	verifier := newFakeVerifier()
	repo := userspostgres.NewUserRepository(db)
	usersUC := users.NewUsecase(repo, ristretto, users.Config{
		BcryptCost: config.AppConfig.BcryptCost,
	})
	authUC := auth.NewUsecase(usersUC, jwtSvc, mailer, verifier, redis, auth.Config{
		OTPMaxAttempts:    5,
		OTPTTL:            5 * time.Minute,
		PasswordResetTTL:  30 * time.Minute,
		BcryptCost:        config.AppConfig.BcryptCost,
		LoginMaxAttempts:  10,
		LoginLockoutTTL:   15 * time.Minute,
		ForgotMaxAttempts: 3,
		ForgotLockoutTTL:  15 * time.Minute,
	})

	return &AuthFixture{
		Auth:     authUC,
		Users:    usersUC,
		Mailer:   mailer,
		Verifier: verifier,
		JWT:      jwtSvc,
	}
}
