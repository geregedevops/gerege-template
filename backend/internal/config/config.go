// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package config

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/viper"
	"templatev27/internal/constants"
)

var AppConfig Config

type Config struct {
	Port        int    `mapstructure:"PORT"`
	Environment string `mapstructure:"ENVIRONMENT"`
	Debug       bool   `mapstructure:"DEBUG"`

	DBPostgreDriver string `mapstructure:"DB_POSTGRE_DRIVER"`
	DBPostgreDsn    string `mapstructure:"DB_POSTGRE_DSN"`
	DBPostgreURL    string `mapstructure:"DB_POSTGRE_URL"`

	DBMaxOpenConns    int `mapstructure:"DB_MAX_OPEN_CONNS"`
	DBMaxIdleConns    int `mapstructure:"DB_MAX_IDLE_CONNS"`
	DBConnMaxLifeMins int `mapstructure:"DB_CONN_MAX_LIFE_MINS"`

	JWTSecret         string `mapstructure:"JWT_SECRET"`
	JWTExpired        int    `mapstructure:"JWT_EXPIRED"`
	JWTIssuer         string `mapstructure:"JWT_ISSUER"`
	JWTRefreshExpired int    `mapstructure:"JWT_REFRESH_EXPIRED"` // хоног

	OTPEmail       string `mapstructure:"OTP_EMAIL"`
	OTPPassword    string `mapstructure:"OTP_PASSWORD"`
	OTPMaxAttempts int    `mapstructure:"OTP_MAX_ATTEMPTS"`

	MailerWorkers   int `mapstructure:"MAILER_WORKERS"`
	MailerQueueSize int `mapstructure:"MAILER_QUEUE_SIZE"`
	MailerRetries   int `mapstructure:"MAILER_RETRIES"`

	BcryptCost int `mapstructure:"BCRYPT_COST"`

	// OTel — OTelExporter хоосон бол tracing идэвхгүй болдог (noop
	// provider). Dev орчинд span-уудыг хэвлэхийн тулд "stdout" гэж тохируул,
	// эсвэл production-д OTEL_EXPORTER_OTLP_ENDPOINT-ийг collector / Jaeger /
	// Tempo / Honeycomb / Datadog endpoint руу заасан "otlp" гэж тохируул.
	OTelExporter    string  `mapstructure:"OTEL_EXPORTER"`
	OTelSampleRatio float64 `mapstructure:"OTEL_SAMPLE_RATIO"`

	REDISHost     string `mapstructure:"REDIS_HOST"`
	REDISPassword string `mapstructure:"REDIS_PASS"`
	REDISExpired  int    `mapstructure:"REDIS_EXPIRED"`

	AllowedOrigins string `mapstructure:"ALLOWED_ORIGINS"`

	// TrustedProxies нь Fiber-ийн X-Forwarded-For / X-Real-IP толгойг
	// итгэлтэйгээр унших ёстой proxy-уудын IP / CIDR жагсаалт (таслалаар
	// тусгаарлагдсан). Empty үед proxy-ийн толгойг үл тоомсорлоно — c.IP()
	// нь TCP peer-ийн IP-г буцаана. Production-д nginx гэх мэт reverse
	// proxy ард ажиллах үед энэ нь шаардлагатай — өөрөөр бол rate limit,
	// audit, access log бүгд proxy-ийн ганц IP харна.
	TrustedProxies string `mapstructure:"TRUSTED_PROXIES"`

	// GeregeCloud Verify (verify.gecloud.mn) — OTP илгээх/шалгах ажлыг
	// гадаад үйлчилгээнд шилжүүлдэг. VerifyAPIKey хоосон бол OTP клиент
	// бүтэх боловч дуудлага бүр "missing api key" алдаа буцаах тул
	// SendOTP/VerifyOTP урсгал тэр даруй амжилтгүй болно — operator-д
	// чимээгүй ажиллахын оронд тодорхой сэрэмжлүүлэг өгөх боллоо.
	VerifyAPIBase string `mapstructure:"VERIFY_API_BASE"`
	VerifyAPIKey  string `mapstructure:"VERIFY_API_KEY"`
	VerifyChannel string `mapstructure:"VERIFY_CHANNEL"`

	// ObservabilityToken нь production-д /metrics ба /swagger/doc.json
	// endpoint-уудыг хамгаалах Bearer токен юм. Хоосон үед production-д
	// эдгээр endpoint 404 буцаана; development-д тэр чигээрээ нээлттэй.
	// Prometheus scraper эсвэл developer Postman үүнийг "Authorization:
	// Bearer <token>" толгойгоор дамжуулна.
	ObservabilityToken string `mapstructure:"OBSERVABILITY_TOKEN"`
}

// TrustedProxiesList нь TRUSTED_PROXIES-г таслалаар тусгаарлан, цэвэрлэсэн
// IP/CIDR жагсаалт болгон буцаана. Empty эсвэл бүгд хоосон үед nil буцаана —
// дуудагч энэ үед Fiber-д proxy итгэлийг идэвхжүүлэхгүй.
func (c *Config) TrustedProxiesList() []string {
	if c.TrustedProxies == "" {
		return nil
	}
	parts := strings.Split(c.TrustedProxies, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// AllowedOriginsList нь CORS origin-уудыг slice болгож буцаана. Зөвхөн хоосон БА орчин production биш үед ["*"] утгыг анхдагчаар авна.
func (c *Config) AllowedOriginsList() []string {
	if c.AllowedOrigins == "" {
		if c.Environment == constants.EnvironmentProduction {
			return nil
		}
		return []string{"*"}
	}
	parts := strings.Split(c.AllowedOrigins, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func InitializeAppConfig() error {
	viper.SetConfigName(".env") // .env файлаас шууд унших боломжийг олгоно
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("internal/config")
	viper.AddConfigPath("/")
	viper.AllowEmptyEnv(true)
	viper.AutomaticEnv()
	// .env файл байхгүй байх нь алдаа БИШ — контейнер / 12-factor орчинд
	// тохиргоог зөвхөн environment-ээс уншина. Зөвхөн жинхэнэ задлан унших
	// (parse) алдааг л буцаана.
	if err := viper.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return constants.ErrLoadConfig
		}
	}

	if err := viper.Unmarshal(&AppConfig); err != nil {
		return constants.ErrParseConfig
	}

	applyDefaults()

	// шалгалт
	if AppConfig.Port == 0 || AppConfig.Environment == "" || AppConfig.JWTSecret == "" || AppConfig.JWTExpired == 0 || AppConfig.JWTIssuer == "" || AppConfig.OTPEmail == "" || AppConfig.OTPPassword == "" || AppConfig.REDISHost == "" || AppConfig.REDISPassword == "" || AppConfig.REDISExpired == 0 || AppConfig.DBPostgreDriver == "" {
		return constants.ErrEmptyVar
	}

	if AppConfig.Port < 1 || AppConfig.Port > 65535 {
		return fmt.Errorf("PORT must be between 1 and 65535, got %d", AppConfig.Port)
	}
	if AppConfig.JWTExpired < 1 || AppConfig.JWTExpired > 720 {
		return fmt.Errorf("JWT_EXPIRED must be between 1 and 720 hours, got %d", AppConfig.JWTExpired)
	}
	if AppConfig.JWTRefreshExpired < 1 || AppConfig.JWTRefreshExpired > 365 {
		return fmt.Errorf("JWT_REFRESH_EXPIRED must be between 1 and 365 days, got %d", AppConfig.JWTRefreshExpired)
	}
	if len(AppConfig.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters (got %d) — HS256 requires 256-bit entropy", len(AppConfig.JWTSecret))
	}
	if AppConfig.REDISExpired < 1 {
		return fmt.Errorf("REDIS_EXPIRED must be at least 1 minute, got %d", AppConfig.REDISExpired)
	}
	if AppConfig.DBMaxOpenConns < 1 || AppConfig.DBMaxIdleConns < 0 || AppConfig.DBMaxIdleConns > AppConfig.DBMaxOpenConns {
		return fmt.Errorf("invalid DB pool config: open=%d idle=%d", AppConfig.DBMaxOpenConns, AppConfig.DBMaxIdleConns)
	}
	if AppConfig.OTPMaxAttempts < 1 {
		return fmt.Errorf("OTP_MAX_ATTEMPTS must be >= 1, got %d", AppConfig.OTPMaxAttempts)
	}
	if AppConfig.BcryptCost < 10 || AppConfig.BcryptCost > 31 {
		return fmt.Errorf("BCRYPT_COST must be between 10 and 31, got %d", AppConfig.BcryptCost)
	}

	switch AppConfig.Environment {
	case constants.EnvironmentDevelopment:
		if AppConfig.DBPostgreDsn == "" {
			return constants.ErrEmptyVar
		}
	case constants.EnvironmentProduction:
		if AppConfig.DBPostgreURL == "" {
			return constants.ErrEmptyVar
		}
		if _, err := url.Parse(AppConfig.DBPostgreURL); err != nil {
			return fmt.Errorf("DB_POSTGRE_URL is not a valid URL: %w", err)
		}
		// secure_system_guide §3.5: production-д DB холболт заавал
		// баталгаажсан TLS-тэй байх ёстой. sslmode=verify-full нь server
		// сертификатыг CA + hostname-ээр шалгаж MITM-аас хамгаална;
		// disable/require/allow/prefer нь сертификатыг шалгахгүй тул
		// production-д хориглоно. (Дотоод сүлжээнд verify-ca-г зөвшөөрнө.)
		if mode := sslModeOf(AppConfig.DBPostgreURL); mode != "verify-full" && mode != "verify-ca" {
			return fmt.Errorf("production DB_POSTGRE_URL must use sslmode=verify-full (got %q) — secure_system_guide §3.5", mode)
		}
		if AppConfig.AllowedOrigins == "" {
			return fmt.Errorf("ALLOWED_ORIGINS must be set in production (comma-separated origins)")
		}
		if err := validateOrigins(AppConfig.AllowedOriginsList()); err != nil {
			return err
		}
		// .env.example-ийн CHANGE_ME placeholder-ууд validation-ийг
		// (length, type) дамжих учир operator санамсаргүй deploy хийвэл
		// "secret" нь жинхэнэ нийтэд мэдэгдэх утга үлдэх эрсдэлтэй.
		// Production-д ийм placeholder-уудыг хатуу татгалзана.
		if err := rejectPlaceholderSecrets(); err != nil {
			return err
		}
	default:
		return fmt.Errorf("ENVIRONMENT must be 'development' or 'production', got %q", AppConfig.Environment)
	}

	return nil
}

// validateOrigins нь ALLOWED_ORIGINS-ын элемент бүрийг scheme://host
// бүхий зөв origin URL мөн эсэхийг шалгана. Production-д wildcard "*"
// эсвэл буруу бичсэн утга нь CORS-ыг чимээгүйгээр сулруулдаг тул
// startup үед шууд унагана.
func validateOrigins(origins []string) error {
	for _, o := range origins {
		if o == "*" {
			return fmt.Errorf("ALLOWED_ORIGINS must not contain wildcard '*' in production")
		}
		u, err := url.Parse(o)
		if err != nil {
			return fmt.Errorf("ALLOWED_ORIGINS contains invalid origin %q: %w", o, err)
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			return fmt.Errorf("ALLOWED_ORIGINS %q must use http(s) scheme (got %q)", o, u.Scheme)
		}
		if u.Host == "" {
			return fmt.Errorf("ALLOWED_ORIGINS %q must include a host", o)
		}
		if u.Path != "" && u.Path != "/" {
			return fmt.Errorf("ALLOWED_ORIGINS %q must not include a path (got %q)", o, u.Path)
		}
	}
	return nil
}

// rejectPlaceholderSecrets нь .env.example-ийн CHANGE_ME ба түүнтэй
// төстэй placeholder-уудыг production-д хүлээн авахаас сэргийлнэ.
// Length check (≥32 байт) эдгээрийг чимээгүй давдаг учир тусдаа guard
// хэрэгтэй. Substring шалгах нь шалтай боловч default placeholder-ыг
// тогтворгүй үлдээхээс илүү — оператор санаатай "CHANGE" гэдэг үг
// орсон жинхэнэ секретээр ажиллахыг хүсвэл нэр өөрчилнө.
func rejectPlaceholderSecrets() error {
	type check struct{ name, val string }
	checks := []check{
		{"JWT_SECRET", AppConfig.JWTSecret},
		{"REDIS_PASS", AppConfig.REDISPassword},
		{"VERIFY_API_KEY", AppConfig.VerifyAPIKey},
		{"OTP_PASSWORD", AppConfig.OTPPassword},
	}
	bad := []string{"CHANGE_ME", "CHANGEME", "PLACEHOLDER", "REPLACE_ME", "TODO_SECRET"}
	for _, c := range checks {
		upper := strings.ToUpper(c.val)
		for _, needle := range bad {
			if strings.Contains(upper, needle) {
				return fmt.Errorf("%s contains placeholder %q — generate a real secret (e.g. `openssl rand -base64 32`) before deploying", c.name, needle)
			}
		}
	}
	return nil
}

// sslModeOf нь Postgres холболтын мөрөөс sslmode утгыг гаргана —
// URL хэлбэр (postgres://...?sslmode=verify-full) болон keyword/DSN
// хэлбэр (host=... sslmode=verify-full) хоёуланг дэмжинэ. sslmode
// байхгүй бол "" буцаана (libpq нь баталгаажуулдаггүй "prefer"-ийг
// өгөгдмөлөөр авах тул production guard үүнийг найдваргүйд тооцно).
func sslModeOf(conn string) string {
	if u, err := url.Parse(conn); err == nil && (u.Scheme == "postgres" || u.Scheme == "postgresql") {
		return strings.ToLower(strings.TrimSpace(u.Query().Get("sslmode")))
	}
	for _, field := range strings.Fields(conn) {
		if k, v, ok := strings.Cut(field, "="); ok && strings.EqualFold(strings.TrimSpace(k), "sslmode") {
			return strings.ToLower(strings.TrimSpace(v))
		}
	}
	return ""
}

// applyDefaults нь сонголттой config утгуудад зохистой анхдагч утгуудыг олгоно.
func applyDefaults() {
	if AppConfig.DBMaxOpenConns == 0 {
		AppConfig.DBMaxOpenConns = 25
	}
	if AppConfig.DBMaxIdleConns == 0 {
		AppConfig.DBMaxIdleConns = 5
	}
	if AppConfig.DBConnMaxLifeMins == 0 {
		AppConfig.DBConnMaxLifeMins = 15
	}
	if AppConfig.OTPMaxAttempts == 0 {
		AppConfig.OTPMaxAttempts = 5
	}
	if AppConfig.MailerWorkers == 0 {
		AppConfig.MailerWorkers = 2
	}
	if AppConfig.MailerQueueSize == 0 {
		AppConfig.MailerQueueSize = 64
	}
	if AppConfig.MailerRetries == 0 {
		AppConfig.MailerRetries = 3
	}
	if AppConfig.BcryptCost == 0 {
		// 12 ≈ 2026 оны үеийн CPU дээр 100–200 мс. bcrypt.DefaultCost нь
		// түүхэн шалтгаанаар одоо ч 10 хэвээр байгаа; үүнийг нэмэгдүүлэв, гэхдээ
		// буруу тохиргоо сервер тээглэхээс сэргийлж bcrypt-ийн өөрийн дээд
		// хэмжээ (31) хүртэл хязгаарлав.
		AppConfig.BcryptCost = 12
	}
	if AppConfig.JWTRefreshExpired == 0 {
		AppConfig.JWTRefreshExpired = 7
	}
	if AppConfig.VerifyChannel == "" {
		AppConfig.VerifyChannel = "email"
	}
	// OTel-ийн sample ratio нь зөвхөн exporter тохируулагдсан БА оператор
	// ratio-г тодорхой зааж өгөөгүй үед 1.0 утгыг анхдагчаар авна. Exporter
	// байхгүй үед ratio нь хамаагүй (noop provider).
	if AppConfig.OTelSampleRatio == 0 && AppConfig.OTelExporter != "" {
		AppConfig.OTelSampleRatio = 1.0
	}
}
