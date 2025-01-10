package main

import (
	"context"
	"database/sql"
	"time"
	
	"github.com/heroiclabs/nakama-common/runtime"

	"territory.com/server/backend"
	"territory.com/server/game"
)

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	initStart := time.Now()

	// Register backend functions
	if err := backend.RegisterAPIs(initializer); err != nil {
		return err
	}

	if err := game.RegisterGame(initializer); err != nil {
		return err
	}

	logger.Info("Module loaded in %dms", time.Since(initStart).Milliseconds())

	return nil
}
