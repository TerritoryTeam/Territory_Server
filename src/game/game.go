package game

import (
	"github.com/heroiclabs/nakama-common/runtime"

	"territory.com/server/game/matches"
)

func RegisterGame(initializer runtime.Initializer) error {
	if err := initializer.RegisterMatch("world", matches.RegisterWorldMatch); err != nil {
		return err
	}

	return nil
}
