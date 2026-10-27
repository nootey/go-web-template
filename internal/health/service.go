package health

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

type Checker interface {
	Name() string
	Check(ctx context.Context) error
}

type DBChecker struct {
	db *sql.DB
}

func NewDBChecker(db *sql.DB) *DBChecker {
	return &DBChecker{db: db}
}

func (d *DBChecker) Name() string                    { return "db" }
func (d *DBChecker) Check(ctx context.Context) error { return d.db.PingContext(ctx) }

type RedisChecker struct {
	rdb *redis.Client
}

func NewRedisChecker(rdb *redis.Client) *RedisChecker {
	return &RedisChecker{rdb: rdb}
}

func (r *RedisChecker) Name() string                    { return "redis" }
func (r *RedisChecker) Check(ctx context.Context) error { return r.rdb.Ping(ctx).Err() }

type Service struct {
	checkers []Checker
	states   map[string]error
	mu       sync.RWMutex
	gauge    metric.Float64Gauge
	logger   *zap.Logger
}

func New(logger *zap.Logger) (*Service, error) {
	gauge, err := otel.GetMeterProvider().Meter("go-web-template").Float64Gauge(
		"health_check_up",
		metric.WithDescription("1 = healthy, 0 = unhealthy"),
	)
	if err != nil {
		return nil, err
	}
	return &Service{
		states: make(map[string]error),
		gauge:  gauge,
		logger: logger,
	}, nil
}

func (s *Service) Add(c Checker) {
	s.checkers = append(s.checkers, c)
}

func (s *Service) Name() string { return "health" }

func (s *Service) Run(ctx context.Context) error {
	s.runChecks(ctx)

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.runChecks(ctx)
		case <-ctx.Done():
			return nil
		}
	}
}

func (s *Service) runChecks(ctx context.Context) {
	for _, c := range s.checkers {
		err := c.Check(ctx)

		s.mu.Lock()
		s.states[c.Name()] = err
		s.mu.Unlock()

		val := 1.0
		if err != nil {
			val = 0.0
			s.logger.Warn("health check failed", zap.String("check", c.Name()), zap.Error(err))
		}
		s.gauge.Record(ctx, val, metric.WithAttributes(attribute.String("check", c.Name())))
	}
}

func (s *Service) Report() Report {
	s.mu.RLock()
	defer s.mu.RUnlock()

	report := Report{Status: StatusOK, Checks: make(map[string]CheckResult, len(s.states))}
	for name, err := range s.states {
		result := CheckResult{Status: StatusOK}
		if err != nil {
			result.Status = StatusError
			report.Status = StatusDegraded
		}
		report.Checks[name] = result
	}
	return report
}
