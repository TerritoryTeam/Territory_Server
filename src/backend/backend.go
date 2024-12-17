package backend

import (
	"net/http"

	"github.com/heroiclabs/nakama-common/runtime"

	"territory.com/server/backend/api"
)

func RegisterAPIs(initializer runtime.Initializer) error {
	// Register the HealthCheck RPC function
	err := initializer.RegisterRpc("HealthCheck", api.RpcHealthCheck)
	if err != nil {
		return err
	}

	if err := initializer.RegisterHttp("/test", api.HttpTest, http.MethodGet); err != nil {
		return err
	}

	return nil
}
