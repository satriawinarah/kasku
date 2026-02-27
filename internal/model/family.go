package model

import "time"

// Family is the top-level group that contains users and wallets.
type Family struct {
	ID        int64
	Name      string
	CreatedAt time.Time
}
