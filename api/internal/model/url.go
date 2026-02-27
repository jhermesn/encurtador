package model

import "time"

type TTL string

const (
	TTL1Hour  TTL = "1h"
	TTL1Day   TTL = "24h"
	TTL1Week  TTL = "168h"
	TTL1Month TTL = "720h"
	TTL1Year  TTL = "8760h"
)

var ValidTTLs = map[TTL]time.Duration{
	TTL1Hour:  time.Hour,
	TTL1Day:   24 * time.Hour,
	TTL1Week:  7 * 24 * time.Hour,
	TTL1Month: 30 * 24 * time.Hour,
	TTL1Year:  365 * 24 * time.Hour,
}

type URL struct {
	ID              uint64    `db:"id"`
	Slug            string    `db:"slug"`
	TargetURL       string    `db:"target_url"`
	PasswordHash    *string   `db:"password_hash"`
	ManageTokenHash string    `db:"manage_token_hash"`
	ExpiresAt       time.Time `db:"expires_at"`
	CreatedAt       time.Time `db:"created_at"`
}

// CachedURL is the payload stored in Redis. It contains everything needed
// to serve a redirect or password gate without hitting MySQL.
type CachedURL struct {
	TargetURL    string `json:"target_url"`
	Protected    bool   `json:"protected"`
	PasswordHash string `json:"password_hash,omitempty"`
}

// ToCached projects a URL into the Redis cache payload.
func (u *URL) ToCached() *CachedURL {
	cached := &CachedURL{
		TargetURL: u.TargetURL,
		Protected: u.PasswordHash != nil,
	}
	if u.PasswordHash != nil {
		cached.PasswordHash = *u.PasswordHash
	}
	return cached
}
