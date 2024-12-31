package api

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/heroiclabs/nakama-common/runtime"
)

func CreateMatchRPC(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	params := make(map[string]interface{})

	if err := json.Unmarshal([]byte(payload), &params); err != nil {
		return "", err
	}

	modulename := "world" // Name with which match handler was registered in InitModule, see example above.

	if matchId, err := nk.MatchCreate(ctx, modulename, params); err != nil {
		return "", err
	} else {
		return matchId, nil
	}
}

func ListAvailableRooms(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	// List all available rooms
	// Return a list of room names
	return "", nil
}
