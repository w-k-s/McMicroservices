package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/w-k-s/McMicroservices/kitchen-service/log"

	dao "github.com/w-k-s/McMicroservices/kitchen-service/internal/persistence"
)

func (a *App) HealthCheck() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		report := make(StatusReport)
		report["database"] = databaseStatusReport(r.Context(), a.pool)

		a.serde.MustSerialize(w, r, report.overallStatus().HttpCode(), report)
	}
}

func databaseStatusReport(ctx context.Context, db *sql.DB) status {
	if err := dao.PingWithBackOff(db); err != nil {
		log.ErrCtx(ctx, err).
			Msg("Ping failed for health check")
		return down
	}
	return up
}

type status bool

const (
	up   status = true
	down status = false
)

func (s status) String() string {
	if s == up {
		return "UP"
	}
	return "DOWN"
}

func (s status) MarshalJSON() ([]byte, error) {
	switch s {
	case up:
		fallthrough
	case down:
		return json.Marshal(s.String())
	default:
		return nil, fmt.Errorf("invalid status: %s", s)
	}
}

func (s *status) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	switch str {
	case "UP":
		*s = up
	case "DOWN":
		*s = down
	default:
		return fmt.Errorf("invalid status: %s", str)
	}
	return nil
}

func (s status) HttpCode() int {
	switch s {
	case up:
		return http.StatusOK
	default:
		return http.StatusInternalServerError
	}
}

type StatusReport map[string]status

func (report StatusReport) overallStatus() status {
	overall := up
	for _, status := range report {
		overall = overall && status
	}
	return overall
}
