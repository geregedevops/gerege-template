// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

// Package ports нь business core (usecase) layer-ийн консьюмер талаас
// зарласан гадаад dependency-ийн контрактуудыг хадгална. Go convention-ийн
// дагуу interface нь хэрэглэгчийн (consumer) талд тодорхойлогддог —
// usecase нь яг ямар method-уудыг шаардаж байгаагаа өөрөө мэддэг тул.
//
// Энэ багц зөвхөн стандарт сангуудаас хамаарна — Redis, GORM, Fiber гэх мэт
// гадны драйвер импортлохгүй. Ингэснээр Clean Architecture-ийн "business
// core нь web/infra layer-аас хамаарахгүй" гэсэн дүрэм хатуу барина.
package ports

import (
	"context"
	"time"
)

// Cache нь usecase давхрагын кэш / TTL-тэй key-value store-ийн порт юм.
// Одоогийн адаптор нь Redis (internal/datasources/caches.redisCache) боловч
// бизнесийн код нь конкретыг мэддэггүй — Memcached, Ristretto-тэй TTL
// wrapper, эсвэл санах ойн fake аль аль нь энэ интерфейсийг хэрэгжүүлэх
// боломжтой.
//
// Цаашид infra-тусгай арга (жнь Redis pipeline, MULTI/EXEC) шаардлагатай
// болвол шинэ usecase-тусгай порт нэмж бизнес-core-ийн контрактыг
// нарийсгана; ийм аргуудыг ENT энд оруулахгүй.
type Cache interface {
	// Set нь value-г JSON болгон marshal хийж, кэшийн нийтлэг өгөгдмөл
	// TTL-тэйгээр key дор бичнэ.
	Set(ctx context.Context, key string, value interface{}) error
	// SetWithTTL нь Set-тэй адил боловч өгөгдмөл TTL-ийг ttl аргументаар
	// орлуулна (атомар SET ... EX ttl).
	SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	// Get нь key дахь JSON-оор decode хийсэн мөрийг буцаана, эсвэл key
	// байхгүй үед сторе-тусгай "not found" sentinel-ийг буцаана (Redis
	// адаптор: redis.Nil).
	Get(ctx context.Context, key string) (string, error)
	// GetDel нь key дахь утгыг авч, тэр key-г атомаар устгана.
	GetDel(ctx context.Context, key string) (string, error)
	// Del нь key-г устгана. Байхгүй key нь алдаа биш.
	Del(ctx context.Context, key string) error
	// Incr нь key дахь бүхэл тоог атомаар нэгээр нэмэгдүүлж, шинэ утгыг
	// буцаана. Key байхгүй бол 1-ээр эхэлнэ.
	Incr(ctx context.Context, key string) (int64, error)
	// Expire нь key дээрх TTL-г (дахин) тогтооно.
	Expire(ctx context.Context, key string, ttl time.Duration) error
	// PTTL нь key-ийн үлдсэн TTL-г миллисекундын нарийвчлалтай буцаана.
	// Key байхгүй бол -2, TTL-гүй бол -1.
	PTTL(ctx context.Context, key string) (time.Duration, error)
}
