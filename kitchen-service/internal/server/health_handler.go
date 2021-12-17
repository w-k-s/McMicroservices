package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	cfg "github.com/w-k-s/McMicroservices/kitchen-service/internal/config"
	dao "github.com/w-k-s/McMicroservices/kitchen-service/internal/persistence"
)

type healthHandler struct {
	Handler
	config *cfg.Config
}

func NewHealthHandler(config *cfg.Config) (healthHandler, error) {
	if config == nil {
		return healthHandler{}, fmt.Errorf("healthHandler received config: nil. non-nil config expected")
	}
	return healthHandler{Handler{}, config}, nil
}

func MustHealthHandler(config *cfg.Config) healthHandler {
	handler, err := NewHealthHandler(config)
	if err != nil {
		log.Fatal(err.Error())
	}
	return handler
}

func (h healthHandler) CheckHealth(w http.ResponseWriter, req *http.Request) {

	report := make(StatusReport)
	report["database"] = h.databaseStatusReport()

	h.MustEncodeJson(w, report, report.overallStatus().HttpCode())
}

func (h healthHandler) databaseStatusReport() status {
	var (
		db  *sql.DB
		err error
	)

	if db, err = sql.Open(
		h.config.Database().DriverName(),
		h.config.Database().ConnectionString()); err != nil {
		log.Printf("Failed to connect to database for health check. Reason: %s", err)
		return down
	}

	db.SetMaxIdleConns(0) // Required, otherwise pinging will result in EOF
	if err = dao.PingWithBackOff(db); err != nil {
		log.Printf("Ping failed for health check. Reason: %s", err)
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
