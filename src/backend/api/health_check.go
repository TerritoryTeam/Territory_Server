package api

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/heroiclabs/nakama-common/runtime"
)

type HealthcheckResponse struct {
	Success bool `json:"success"`
}

func RpcHealthCheck(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	logger.Debug("Healthcheck request")

	response := &HealthcheckResponse{Success: true}

	out, err := json.Marshal(response)
	if err != nil {
		logger.Error("Error encoding response: %v", err)
		return "", runtime.NewError("Error encoding response", 500)
	}

	return string(out), nil
}
