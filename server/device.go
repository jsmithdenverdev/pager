package main

type device struct {
	auditable
	ID string `json:"id" db:"id"`
}
