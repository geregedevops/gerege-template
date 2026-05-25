// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/time/rate"
	"gorm.io/gorm"
	docs "templatev27/docs" // swagger тодорхойлолт, swaggo-оор init үед бүртгэгддэг
	"templatev27/internal/business/usecases/auth"
	"templatev27/internal/business/usecases/users"
	"templatev27/internal/config"
	"templatev27/internal/constants"
	"templatev27/internal/datasources/caches"
	"templatev27/internal/datasources/drivers"
	userspostgres "templatev27/internal/datasources/repositories/postgres/users"
	V1Handler "templatev27/internal/http/handlers/v1"
	"templatev27/internal/http/middlewares"
	"templatev27/internal/http/routes"
	"templatev27/pkg/jwt"
	"templatev27/pkg/logger"
	"templatev27/pkg/mailer"
	"templatev27/pkg/observability"
	"templatev27/pkg/verify"
)

const serviceName = "gerege-template"

type App struct {
	fiber           *fiber.App
	db              *gorm.DB
	redisCache      caches.RedisCache
	asyncMailer     *mailer.AsyncOTPMailer
	tracerShutdown  observability.Shutdown
	authRateLimiter *middlewares.RateLimiter
}

func NewApp() (*App, error) {
	// Tracer-ийг эхэлд тохируулна — ингэснээр дараагийн тохиргооноос (DB холболт,
	// migration шалгалт гэх мэт) ялгарах span-ууд зөв provider руу очно.
	shutdownTracer, err := observability.SetupTracing(context.Background(), observability.TracingConfig{
		ServiceName: serviceName,
		Environment: config.AppConfig.Environment,
		Exporter:    config.AppConfig.OTelExporter,
		SampleRatio: config.AppConfig.OTelSampleRatio,
	})
	if err != nil {
		return nil, fmt.Errorf("setup tracing: %w", err)
	}

	// өгөгдлийн сангуудыг тохируулах
	conn, err := drivers.SetupGORMPostgres()
	if err != nil {
		return nil, err
	}
	// DB pool-ийн бодит статистикийг /metrics-ээр гаргана — provider нь GORM
	// handle-аас авсан түүхий *sql.DB-г буцаана.
	observability.RegisterDBStatsProvider(func() *sql.DB {
		sqlDB, dbErr := conn.DB()
		if dbErr != nil {
			return nil
		}
		return sqlDB
	})

	// jwt сервис
	jwtService := jwt.NewJWTServiceWithRefresh(
		config.AppConfig.JWTSecret,
		config.AppConfig.JWTIssuer,
		config.AppConfig.JWTExpired,
		config.AppConfig.JWTRefreshExpired,
	)

	// кэш
	redisCache := caches.NewRedisCache(config.AppConfig.REDISHost, 0, config.AppConfig.REDISPassword, time.Duration(config.AppConfig.REDISExpired))
	ristrettoCache, err := caches.NewRistrettoCache()
	if err != nil {
		return nil, fmt.Errorf("failed to create ristretto cache: %w", err)
	}

	// mailer — синхрон SMTP илгээгчийг асинхрон дараалалд (queue) боож,
	// OTP илгээх хоцролтыг HTTP хүсэлтийн замаас гаргана.
	syncMailer := mailer.NewOTPMailer(config.AppConfig.OTPEmail, config.AppConfig.OTPPassword)
	asyncMailer := mailer.NewAsyncOTPMailer(
		syncMailer,
		config.AppConfig.MailerWorkers,
		config.AppConfig.MailerQueueSize,
		config.AppConfig.MailerRetries,
		time.Second,
	)

	// router + глобал middleware-г тохируулах
	app := setupRouter()

	// auth middleware — хүчинтэй токентой хэрэглэгч endpoint-д хандах боломжтой.
	// redisCache-г дамжуулсан тул middleware нь ChangePassword / ResetPassword-оор
	// бичигдсэн нууц үг солих хугацааны хязгаарыг (cutoff) баримтлаж чадна.
	authMiddleware := middlewares.NewAuthMiddleware(jwtService, redisCache, false)

	// Дэд бүтцийн (infrastructure) endpoint-ууд (/api бүлгээс гадуур)
	healthHandler := V1Handler.NewHealthHandler(conn, redisCache.Client())
	app.Get("/health", healthHandler.Health)
	app.Get("/ready", healthHandler.Ready)
	// /metrics — Prometheus exposition. promhttp нь net/http handler бөгөөд
	// adaptor middleware-ээр дамжуулан Fiber руу холбогддог.
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))
	// OpenAPI тодорхойлолт — `make swag` нь godoc annotation-уудаас docs/
	// багцыг үүсгэдэг. gofiber/swagger нь Fiber v2-д зориулагдсан тул
	// (handler нь *fiber.Ctx авдаг) Fiber v3-д runtime panic үүсгэдэг —
	// иймд spec-ийг Fiber v3 native-аар JSON хэлбэрээр үйлчилнэ. Уг JSON-ыг
	// Swagger UI / Postman / VS Code-д шууд ачаалж болно. (Суулгасан
	// интерактив UI хэрэгтэй бол Fiber v3-тэй нийцэх swagger handler нэмнэ.)
	app.Get("/swagger/doc.json", func(c fiber.Ctx) error {
		c.Set("Content-Type", "application/json")
		return c.SendString(docs.SwaggerInfo.ReadDoc())
	})

	// Хязгаарлагдсан контекстуудыг (bounded context) угсарна. Users нь
	// identity CRUD-г эзэмшиж, Auth нь credential / session урсгалыг эзэмшинэ;
	// Auth нь хэрэглэгчийн бичлэгийг унших/бичихдээ Users-ээс хамаардаг.
	userRepo := userspostgres.NewUserRepository(conn)
	usersUC := users.NewUsecase(userRepo, ristrettoCache, users.Config{
		BcryptCost: config.AppConfig.BcryptCost,
	})
	// GeregeCloud Verify клиент — OTP илгээх/шалгах ажлыг алсын үйлчилгээнд
	// шилжүүлнэ. VerifyAPIKey хоосон бол клиент бүтэх боловч дуудлага бүр
	// "missing api key" алдаа буцаах тул operator-д чимээгүй буруу тохиргоо
	// үлдэхгүй.
	verifyClient := verify.NewClient(
		config.AppConfig.VerifyAPIBase,
		config.AppConfig.VerifyAPIKey,
		config.AppConfig.VerifyChannel,
	)
	authUC := auth.NewUsecase(usersUC, jwtService, asyncMailer, verifyClient, redisCache, auth.Config{
		OTPMaxAttempts:    config.AppConfig.OTPMaxAttempts,
		OTPTTL:            time.Duration(config.AppConfig.REDISExpired) * time.Minute,
		PasswordResetTTL:  30 * time.Minute,
		BcryptCost:        config.AppConfig.BcryptCost,
		LoginMaxAttempts:  10,
		LoginLockoutTTL:   15 * time.Minute,
		ForgotMaxAttempts: 3,
		ForgotLockoutTTL:  15 * time.Minute,
	})

	// Нэргүй /auth гадаргуун дээр IP тус бүрт минутанд 5 хүсэлт зөвшөөрнө.
	// App нь түүнийг эзэмшдэг тул фоны цэвэрлэгээний goroutine-ийг
	// graceful shutdown (эвсэг унтраалт) үед зогсоож болно.
	authRateLimiter := middlewares.NewRateLimiter(rate.Limit(5.0/60.0), 5)

	// API Route-ууд
	api := app.Group("/api")
	api.Get("/", routes.RootHandler)
	routes.NewAuthRoute(api, authUC, authMiddleware, authRateLimiter).Routes()
	routes.NewUsersRoute(api, usersUC, authMiddleware).Routes()

	return &App{
		fiber:           app,
		db:              conn,
		redisCache:      redisCache,
		asyncMailer:     asyncMailer,
		tracerShutdown:  shutdownTracer,
		authRateLimiter: authRateLimiter,
	}, nil
}

func (a *App) Run() (err error) {
	srvLog := logger.WithFields(logger.Fields{constants.LoggerCategory: constants.LoggerCategoryServer})

	addr := fmt.Sprintf(":%d", config.AppConfig.Port)
	go func() {
		srvLog.Infof("success to listen and serve on %s", addr)
		if listenErr := a.fiber.Listen(addr); listenErr != nil && !errors.Is(listenErr, http.ErrServerClosed) {
			srvLog.Fatalf("Failed to listen and serve: %+v", listenErr)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	srvLog.Info("shutdown server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Шинэ холболт хүлээж авахаа болиод, явагдаж буй хүсэлтүүдийг гүйцээнэ.
	if shutdownErr := a.fiber.ShutdownWithContext(ctx); shutdownErr != nil {
		return fmt.Errorf("error when shutdown server: %v", shutdownErr)
	}

	// Rate limiter-ийн цэвэрлэгээний goroutine-ийг зогсооно. Shutdown-ийн
	// дараа дуудахад аюулгүй — middleware руу шинэ хүсэлт ирэхгүй.
	if a.authRateLimiter != nil {
		a.authRateLimiter.Stop()
	}

	// хамаарлуудыг устгахаас өмнө асинхрон mailer-ийн дарааллыг гүйцээж,
	// явагдаж буй OTP имэйлүүд хүргэгдэх боломжтой болгоно.
	if a.asyncMailer != nil {
		if err := a.asyncMailer.Shutdown(ctx); err != nil {
			srvLog.Errorf("mailer shutdown incomplete: %v", err)
		}
	}

	// өгөгдлийн сангийн холболтыг хаах
	if sqlDB, dbErr := a.db.DB(); dbErr == nil {
		if closeErr := sqlDB.Close(); closeErr != nil {
			srvLog.Errorf("error closing database: %v", closeErr)
		}
	}

	// redis холболтыг хаах
	if err := a.redisCache.Close(); err != nil {
		srvLog.Errorf("error closing redis: %v", err)
	}

	// batch exporter-ийн буферлэж буй бүх span-уудыг flush хийнэ — HTTP сервер
	// хүсэлт хүлээж авахаа больсны дараа боловч процесс гарахаас өмнө ажиллах
	// ёстой, эс бөгөөс явагдаж буй trace-ийн төгсгөлийн хэсэг алдагдана.
	if a.tracerShutdown != nil {
		if err := a.tracerShutdown(ctx); err != nil {
			srvLog.Errorf("tracer shutdown incomplete: %v", err)
		}
	}

	srvLog.Info("server exiting")
	return
}

// setupRouter нь Fiber app-ыг бүтээж, глобал middleware стекийг суулгана.
// Дараалал чухал: эхэлд tracing — ингэснээр RequestIDMiddleware түүнийг
// logger context руу холбохоос өмнө span context (trace_id) тогтоогддог;
// стекийн доош ялгарах span-ууд (DB, Redis) автоматаар серверийн span-ийн
// дэд (child) болдог.
func setupRouter() *fiber.App {
	fiberCfg := fiber.Config{
		// Framework түвшний body-ийн дээд хязгаар — хамгаалалтын эхний шугам.
		// Route тус бүрийн илүү чанга хязгаарыг BodySizeLimitMiddleware-ээр тавина.
		BodyLimit: int(middlewares.DefaultBodyMaxBytes),
		// Төвлөрсөн алдааны handler: handler-ийн буцаасан аливаа алдаа (эсвэл
		// дээд талд сэргээгдсэн panic) энд цуглуулагдаж, нэгдсэн BaseResponse
		// дугтуй (envelope)-аар дүрслэгдэнэ.
		ErrorHandler: func(c fiber.Ctx, err error) error {
			return V1Handler.RespondWithError(c, err)
		},
	}
	// Reverse proxy ард байх үед (nginx, ALB, Cloudflare г.м.) X-Forwarded-For-ийг
	// итгэлтэйгээр уншихын тулд TRUSTED_PROXIES-ийг тохируул. Тохиргоогүй үед
	// Fiber нь спуфинг хийсэн толгойг үл тоомсорлоод TCP peer-ийн IP-г буцаана —
	// энэ нь dev-д зөв (proxy байхгүй), харин production-д хууль ёсны клиентүүд
	// бүгд proxy-ийн ганц IP харагдах тул rate limit / audit / access log
	// эвдэрнэ. Operator оруулсан үед EnableIPValidation ороод header утга нь
	// үнэхээр IP байгаа эсэхийг шалгана.
	if proxies := config.AppConfig.TrustedProxiesList(); len(proxies) > 0 {
		fiberCfg.TrustProxy = true
		fiberCfg.TrustProxyConfig = fiber.TrustProxyConfig{Proxies: proxies}
		fiberCfg.ProxyHeader = fiber.HeaderXForwardedFor
		fiberCfg.EnableIPValidation = true
	}
	app := fiber.New(fiberCfg)

	app.Use(middlewares.TracingMiddleware(serviceName))
	app.Use(middlewares.RequestIDMiddleware())
	app.Use(middlewares.MetricsMiddleware())
	app.Use(middlewares.SecurityHeadersMiddleware())
	app.Use(middlewares.CORSMiddleware())
	app.Use(middlewares.BodySizeLimitMiddleware(middlewares.DefaultBodyMaxBytes))
	app.Use(middlewares.AccessLogMiddleware())
	// Хүсэлт бүрт deadline — гацсан handler/query холболтыг хэт удаан
	// эзлэхээс сэргийлнэ (secure_system_guide §5.3). tracing/request-id-ийн
	// дараа байрлуулсан тул context-ийн утгууд хадгалагдана.
	app.Use(middlewares.TimeoutMiddleware(middlewares.DefaultRequestTimeout))

	return app
}
