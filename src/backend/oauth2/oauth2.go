package oauth2

import (
	"context"
)

type ProvidedUser struct {
	// ID is the unique ID of thirt-party platform.
	ID string
	// Email is the email of thirt-party platform.
	Email string
	// Username is the username of thirt-party platform.
	Username *string
}

type OauthProvider struct {
	// Name is OAuth platform name.
	Name OauthName
	// FetchUser is a function to fetch user information from OAuth platform.
	FetchUser func(ctx context.Context, token *oauth2.Token) (ProvidedUser, error)
}

type OauthName string

const (
	Github OauthName = "github"
)
