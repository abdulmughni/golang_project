package models

import "github.com/golang-jwt/jwt/v5"

type QueryParamToken struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
	jwt.RegisteredClaims
}
