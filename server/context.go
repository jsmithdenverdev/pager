package main

// pagerContextKey is a unique identifier for a pager request context
type pagerContextKey struct{}

// pagerContext represents request scoped values accessible from graphql
// resolvers.
type pagerContext struct {
	User string `json:"user"`
}
