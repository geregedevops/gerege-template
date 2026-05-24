// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

// Package migration нь cmd/migration-ийн ард байрлах туршиж болох
// сан юм. cmd/migration/main.go дахь CLI нь одоо энэ package-ийн
// нимгэн бүрхүүл болсон — config ачаалах + flag задлах + db холбох —
// тиймээс idempotency / advisory-lock / нэг файлд нэг транзакцийн
// зан төлөвийг binary ажиллуулалгүйгээр integration тестэд шалгаж болно.
package migration

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"templatev27/internal/constants"
	"templatev27/internal/datasources/records"
	"templatev27/pkg/logger"

	"gorm.io/gorm"
)

// AdvisoryLockID нь pg_advisory_lock-той хамт ашиглагддаг дурын
// 64-бит бүхэл тоо бөгөөд хоёр migration runner нэг файлыг зэрэг
// хэрэгжүүлэхээс сэргийлдэг. CI smoke тест lock төлөвийг шалгаж
// болохын тулд export хийгдсэн.
const AdvisoryLockID = 947328461230

// Runner нь schema_migrations хүснэгтэд хэрэгжсэн төлөвийг хянахын
// зэрэгцээ Postgres DB-д SQL migration файлуудыг хэрэгжүүлэх /
// буцаах үйлдлийг гүйцэтгэнэ. DB handle нь GORM; доор байрлах
// түүхий *sql.DB нь нэг файлд нэг транзакц болон schema_migrations-ийн
// бүртгэлийг удирддаг (анхны sqlx суурьтай runner-ээс семантик нь
// өөрчлөгдөөгүй).
type Runner struct {
	gormDB *gorm.DB
	db     *sql.DB
	dir    string
	// log нь тестүүдэд no-op sink сольж тавих боломж олгоно; nil бол
	// төслийн нийтлэг logger руу буцна.
	log func(msg string, fields logger.Fields)
}

// New нь `dir`-ээс migration файлуудыг уншдаг Runner-г бүтээнэ.
func New(db *gorm.DB, dir string) *Runner {
	sqlDB, _ := db.DB()
	return &Runner{gormDB: db, db: sqlDB, dir: dir}
}

// SetLogger нь өгөгдмөл logger sink-г дарж бичнэ. Тестийн үед runner-ийг
// чимээгүй болгохын тулд no-op дамжуул.
func (r *Runner) SetLogger(fn func(string, logger.Fields)) { r.log = fn }

// DB нь migration хийсний дараах төлөвийг шалгах шаардлагатай дуудагчдад
// (жишээ нь schema_migrations-г шууд асуудаг integration тест) зориулж
// үндсэн GORM handle-г илчилнэ. Production код үүнийг ашиглах ёсгүй —
// Up/Down/AutoMigrate method-уудыг ашигла.
func (r *Runner) DB() *gorm.DB { return r.gormDB }

// AutoMigrate нь төслийн model-уудад зориулж GORM-ийн idempotent
// schema sync-г ажиллуулна. Энэ нь SQL файлуудыг нөхдөг: AutoMigrate нь
// model-оос гаралтай баганануудыг үүсгэдэг/шинэчилдэг бол *.up.sql
// файлууд нь AutoMigrate илэрхийлж чадахгүй хэсгүүдийг (uuid_generate_v4()
// id-ийн өгөгдмөл утгыг дэмждэг uuid-ossp өргөтгөл, хэсэгчилсэн давтагдашгүй
// индексүүд болон email/username хэвийн болголтыг) хангадаг. Дахин дахин
// дуудахад аюулгүй.
func (r *Runner) AutoMigrate(ctx context.Context) error {
	r.info("running gorm auto-migrate", logger.Fields{constants.LoggerCategory: constants.LoggerCategoryMigration})
	if err := r.gormDB.WithContext(ctx).AutoMigrate(&records.Users{}); err != nil {
		return fmt.Errorf("auto-migrate: %w", err)
	}
	r.info("gorm auto-migrate success", logger.Fields{constants.LoggerCategory: constants.LoggerCategoryMigration})
	return nil
}

func (r *Runner) info(msg string, fields logger.Fields) {
	if r.log != nil {
		r.log(msg, fields)
		return
	}
	logger.Info(msg, fields)
}

// Up нь бүх *.up.sql файлыг лексикографийн дарааллаар хэрэгжүүлнэ.
// schema_migrations-д аль хэдийн байгаа файлуудыг алгасдаг тул дахин
// ажиллуулалт idempotent байна. Файл бүр өөрийн statement болон
// schema_migrations мөрийг нэг транзакцид commit хийнэ.
func (r *Runner) Up(ctx context.Context) error {
	r.info("running migration [up]", logger.Fields{constants.LoggerCategory: constants.LoggerCategoryMigration})

	if err := r.ensureMigrationsTable(ctx); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	return r.withAdvisoryLock(ctx, func() error {
		files, err := r.listFiles("up")
		if err != nil {
			return err
		}
		applied, err := r.loadApplied(ctx)
		if err != nil {
			return err
		}

		for _, file := range files {
			name := filepath.Base(file)
			if applied[name] {
				r.info("skipping already-applied migration", logger.Fields{
					constants.LoggerCategory: constants.LoggerCategoryMigration,
					constants.LoggerFile:     name,
				})
				continue
			}
			r.info("applying migration", logger.Fields{
				constants.LoggerCategory: constants.LoggerCategoryMigration,
				constants.LoggerFile:     name,
			})
			if err := r.applyFile(ctx, file, name, true); err != nil {
				return err
			}
		}
		r.info("migration [up] success", logger.Fields{constants.LoggerCategory: constants.LoggerCategoryMigration})
		return nil
	})
}

// Down нь бүх *.down.sql файлыг лексикографийн ЭСРЭГ дарааллаар
// хэрэгжүүлнэ (хожуу migration эхэлж буцдаг). Амжилттай down бүр
// тохирох schema_migrations мөрийг устгана.
func (r *Runner) Down(ctx context.Context) error {
	r.info("running migration [down]", logger.Fields{constants.LoggerCategory: constants.LoggerCategoryMigration})

	if err := r.ensureMigrationsTable(ctx); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	return r.withAdvisoryLock(ctx, func() error {
		files, err := r.listFiles("down")
		if err != nil {
			return err
		}
		sort.Sort(sort.Reverse(sort.StringSlice(files)))

		for _, file := range files {
			name := filepath.Base(file)
			r.info("reverting migration", logger.Fields{
				constants.LoggerCategory: constants.LoggerCategoryMigration,
				constants.LoggerFile:     name,
			})
			if err := r.applyFile(ctx, file, deriveUpName(name), false); err != nil {
				return err
			}
		}
		r.info("migration [down] success", logger.Fields{constants.LoggerCategory: constants.LoggerCategoryMigration})
		return nil
	})
}

func (r *Runner) ensureMigrationsTable(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			name        TEXT PRIMARY KEY,
			applied_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
	return err
}

// withAdvisoryLock нь зэрэгцээ runner-уудыг (CI + инженерийн зөөврийн
// компьютер) дараалалд оруулж, нэг удаад зөвхөн нэг нь schema-г
// өөрчилж чадахаар болгоно.
func (r *Runner) withAdvisoryLock(ctx context.Context, fn func() error) error {
	if _, err := r.db.ExecContext(ctx, `SELECT pg_advisory_lock($1)`, AdvisoryLockID); err != nil {
		return fmt.Errorf("acquire advisory lock: %w", err)
	}
	defer func() {
		if _, err := r.db.ExecContext(ctx, `SELECT pg_advisory_unlock($1)`, AdvisoryLockID); err != nil {
			logger.Error("failed to release migration advisory lock", logger.Fields{
				constants.LoggerCategory: constants.LoggerCategoryMigration,
				"error":                  err.Error(),
			})
		}
	}()
	return fn()
}

func (r *Runner) listFiles(action string) ([]string, error) {
	files, err := filepath.Glob(filepath.Join(r.dir, fmt.Sprintf("*.%s.sql", action)))
	if err != nil {
		return nil, errors.New("glob migration files")
	}
	sort.Strings(files)
	return files, nil
}

func (r *Runner) loadApplied(ctx context.Context) (map[string]bool, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT name FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("load applied migrations: %w", err)
	}
	defer func() { _ = rows.Close() }()

	applied := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		applied[name] = true
	}
	return applied, rows.Err()
}

// applyFile нь migration SQL файлыг schema_migrations-ийн бүртгэлийн
// бичилттэй хамт нэг транзакцид ажиллуулдаг — ингэснээр файлын дунд
// гацах нь хэсэгчилсэн бичлэг үлдээдэггүй.
func (r *Runner) applyFile(ctx context.Context, file, upName string, isUp bool) error {
	// #nosec G304 — файлын замууд нь хүсэлтийн оролтоос биш, хөгжүүлэгчийн
	// хяналт дахь migrations директороос ирдэг.
	data, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("read %s: %w", file, err)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, string(data)); err != nil {
		return fmt.Errorf("exec %s: %w", filepath.Base(file), err)
	}

	if isUp {
		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations(name) VALUES ($1) ON CONFLICT DO NOTHING`, upName); err != nil {
			return fmt.Errorf("record migration %s: %w", upName, err)
		}
	} else {
		if _, err := tx.ExecContext(ctx, `DELETE FROM schema_migrations WHERE name = $1`, upName); err != nil {
			return fmt.Errorf("forget migration %s: %w", upName, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit %s: %w", filepath.Base(file), err)
	}
	return nil
}

// deriveUpName нь "*.down.sql" файлын нэрийг түүний "*.up.sql"
// хослол болгон хувиргадаг бөгөөд migration-ууд schema_migrations-д
// яг ийм байдлаар түлхүүрлэгддэг.
func deriveUpName(downName string) string {
	const suffix = ".down.sql"
	if len(downName) > len(suffix) && downName[len(downName)-len(suffix):] == suffix {
		return downName[:len(downName)-len(suffix)] + ".up.sql"
	}
	return downName
}
