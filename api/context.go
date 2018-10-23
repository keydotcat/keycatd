package api

import (
	"context"

	"github.com/keydotcat/keycatd/managers"
	"github.com/keydotcat/keycatd/models"
)

type contextType int

const (
	contextUserKey    = contextType(0)
	contextTeamKey    = contextType(iota)
	contextVaultKey   = contextType(iota)
	contextSessionKey = contextType(iota)
	contextCsrfKey    = contextType(iota)
)

func ctxAddUser(ctx context.Context, u *models.User) context.Context {
	return context.WithValue(ctx, contextUserKey, u)
}

func ctxGetUser(ctx context.Context) *models.User {
	d, ok := ctx.Value(contextUserKey).(*models.User)
	if !ok {
		panic("No user defined in context")
	}
	return d
}

func ctxAddTeam(ctx context.Context, u *models.Team) context.Context {
	return context.WithValue(ctx, contextTeamKey, u)
}

func ctxGetTeam(ctx context.Context) *models.Team {
	d, ok := ctx.Value(contextTeamKey).(*models.Team)
	if !ok {
		panic("No team defined in context")
	}
	return d
}

func ctxAddVault(ctx context.Context, u *models.Vault) context.Context {
	return context.WithValue(ctx, contextVaultKey, u)
}

func ctxGetVault(ctx context.Context) *models.Vault {
	d, ok := ctx.Value(contextVaultKey).(*models.Vault)
	if !ok {
		panic("No vault defined in context")
	}
	return d
}

func ctxAddSession(ctx context.Context, u *managers.Session) context.Context {
	return context.WithValue(ctx, contextSessionKey, u)
}

func ctxGetSession(ctx context.Context) *managers.Session {
	d, ok := ctx.Value(contextSessionKey).(*managers.Session)
	if !ok {
		panic("No session defined in context")
	}
	return d
}

func ctxAddCsrf(ctx context.Context, u string) context.Context {
	return context.WithValue(ctx, contextCsrfKey, u)
}

func ctxGetCsrf(ctx context.Context) string {
	d, ok := ctx.Value(contextCsrfKey).(string)
	if !ok {
		panic("No csrf defined in context")
	}
	return d
}
