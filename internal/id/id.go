package id

import "github.com/google/uuid"

// New returns a new UUIDv4 string.
func New() string { return uuid.NewString() }
