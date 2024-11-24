package auth

import (
	"context"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
)

var secrets struct {
	BearerToken string
}

type Data struct {
	Username string
}

//encore:authhandler
func AuthHandler(ctx context.Context, token string) (auth.UID, *Data, error) {
	if token != secrets.BearerToken {
		return "", nil, &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "invalid token",
		}
	}
	return "user", &Data{Username: "user"}, nil
}
