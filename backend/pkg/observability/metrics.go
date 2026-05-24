// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

// Package observability нь бизнес package-ууд болон collector бүртгэдэг HTTP
// middleware-ийн хооронд import цикл үүсгэлгүйгээр аливаа давхаргаас дуудаж
// болох Prometheus metric туслахуудыг ил гаргадаг.
//
// Collector-уудыг энд init() үед нэг удаа бүртгэдэг; дуудагчид үйл явдал
// тэмдэглэхдээ доорх Observe* функцуудыг ашиглана. HTTP давхарга нь өөрийн
// хүсэлтийн хүрээнд (request-scoped) collector-уудыг тусад нь бүртгэдэг.
package observability

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	cacheOpsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_operations_total",
			Help: "Cache operations by layer (ristretto|redis), operation, and result (hit|miss|error|ok).",
		},
		[]string{"layer", "op", "result"},
	)

	mailerOpsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mailer_operations_total",
			Help: "OTP mailer outcomes: sent, failed, queue_full.",
		},
		[]string{"result"},
	)

	dbPoolOpen = prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "db_pool_open_connections",
			Help: "Open Postgres connections (idle + in use)",
		},
		func() float64 { return float64(currentDBStats().OpenConnections) },
	)
	dbPoolInUse = prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "db_pool_in_use_connections",
			Help: "Postgres connections currently in use",
		},
		func() float64 { return float64(currentDBStats().InUse) },
	)
	dbPoolWait = prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "db_pool_wait_count_total",
			Help: "Cumulative connections that waited for a free slot",
		},
		func() float64 { return float64(currentDBStats().WaitCount) },
	)
)

func init() {
	prometheus.MustRegister(cacheOpsTotal, mailerOpsTotal, dbPoolOpen, dbPoolInUse, dbPoolWait)
}

// dbStatsProvider нь эхлэх үед холбогддог бөгөөд ингэснээр GaugeFunc callback-ууд
// server package-г import хийлгүйгээр амьд sql.DBStats-г унших боломжтой болно.
// Энэ нь gormDB.DB()-ээр GORM handle-аас авсан түүхий *sql.DB-г буцаана.
var dbStatsProvider func() *sql.DB

type dbStatsSnapshot struct {
	OpenConnections int
	InUse           int
	WaitCount       int64
}

func currentDBStats() dbStatsSnapshot {
	if dbStatsProvider == nil {
		return dbStatsSnapshot{}
	}
	db := dbStatsProvider()
	if db == nil {
		return dbStatsSnapshot{}
	}
	s := db.Stats()
	return dbStatsSnapshot{
		OpenConnections: s.OpenConnections,
		InUse:           s.InUse,
		WaitCount:       s.WaitCount,
	}
}

// RegisterDBStatsProvider-г эхлэх үед амьд *sql.DB-г (gormDB.DB()-ээс) өгдөг
// provider-ийн хамт нэг удаа дуудах ёстой бөгөөд ингэснээр pool-статистикийн
// gauge-ууд scrape бүрт түүнийг унших боломжтой болно.
func RegisterDBStatsProvider(provider func() *sql.DB) {
	dbStatsProvider = provider
}

// ObserveCacheOp нь нэг кэш үйлдлийн үр дүнг тэмдэглэнэ.
//
//	layer:  "ristretto" | "redis"
//	op:     "get" | "set" | "del"
//	result: "hit" | "miss" | "ok" | "error"
func ObserveCacheOp(layer, op, result string) {
	cacheOpsTotal.WithLabelValues(layer, op, result).Inc()
}

// ObserveMailerOp нь нэг mailer-ийн үр дүнг тэмдэглэнэ.
//
//	result: "sent" | "failed" | "queue_full"
func ObserveMailerOp(result string) {
	mailerOpsTotal.WithLabelValues(result).Inc()
}
