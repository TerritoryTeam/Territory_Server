package backend

import (
	"net/http"

	"github.com/heroiclabs/nakama-common/runtime"

	"territory.com/server/backend/api"
)

func RegisterAPIs(initializer runtime.Initializer) error {

	// Register the AuthenticateCustom function
	if err := initializer.RegisterBeforeAuthenticateCustom(api.BeforeAuthenticateCustom); err != nil {
		return err
	}

	// Register the HealthCheck RPC function

	if err := initializer.RegisterRpc("HealthCheck", api.RpcHealthCheck); err != nil {
		return err
	}

	if err := initializer.RegisterHttp("/test", api.HttpTest, http.MethodGet); err != nil {
		return err
	}

	return nil
}
