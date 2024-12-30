package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/heroiclabs/nakama-common/runtime"
)

func CreateMatchRPC(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	params := make(map[string]interface{})

	if err := json.Unmarshal([]byte(payload), &params); err != nil {
		return "", err
	}

	modulename := "lobby" // Name with which match handler was registered in InitModule, see example above.

	if matchId, err := nk.MatchCreate(ctx, modulename, params); err != nil {
		return "", err
	} else {
		return matchId, nil
	}
}

type TestResponse struct {
	Success bool `json:"success"`
}

func HttpTest(w http.ResponseWriter, r *http.Request) {
	response := &TestResponse{Success: true}

	out, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}
