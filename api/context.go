package models

import (
	"context"

	"github.com/keydotcat/backend/models"
)

const (
	contextUserKey  = 0
	contextTeamKey  = 1
	contextVaultKey = 2
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
