package api

import (
	"context"
	"database/sql"
	"io"
	"net/http"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
)

func BeforeAuthenticateCustom(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, in *api.AuthenticateCustomRequest) (*api.AuthenticateCustomRequest, error) {
	logger.Info("Before Authenticate Custom")

	// default to use github as authenticate provider
	const apiUrl = "https://api.github.com";

	// Construct a payload to send to the third-party API
	const payload = JSON.stringify({
		id: in.account.id
	});
	
	const response = nk.httpRequest(apiUrl, 'post', { 'content-type': 'application/json' }, JSON.stringify(payload));
	if (response.code > 299) {
		logger.error(`API error: ${response.body}`);
		return null
	}
	
	const userInfo = JSON.parse(response.body);
	if (!userInfo.userId || !userInfo.username) {
		logger.error(`invalid API response: ${response.body}`)
		return null;
	}
	
	// Update the incoming authenticate request with the new user ID and username
	in.account.id = userInfo.userId;
	in.username = userInfo.username;
	
	return in, nil;
}
