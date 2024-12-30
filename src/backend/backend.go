package backend

import (
	"net/http"

	"github.com/heroiclabs/nakama-common/runtime"

	"territory.com/server/backend/api"
	"territory.com/server/backend/game"
)

func RegisterAPIs(initializer runtime.Initializer) error {

	initializer.RegisterHttp("/test", api.HttpTest, http.MethodGet)

	// Register the HealthCheck RPC function
	initializer.RegisterRpc("HealthCheck", api.RpcHealthCheck)

	if err := initializer.RegisterMatch("world", game.RegisterWorldMatch); err != nil {
		return err
	}

	// Register the CreateMatch RPC function
	initializer.RegisterRpc("CreateMatch", api.CreateMatchRPC)

	return nil
}
