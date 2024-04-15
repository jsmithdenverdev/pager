package schema

import (
	"fmt"

	"github.com/graphql-go/graphql"
	"github.com/jsmithdenverdev/pager/models"
)

type pageInfo struct {
	HasNextPage     bool   `json:"hasNextPage"`
	HasPreviousPage bool   `json:"hasPreviousPage"`
	StartCursor     string `json:"startCursor"`
	EndCursor       string `json:"endCursor"`
}

type edge[T models.Identity] struct {
	Cursor string `json:"cursor"`
	Node   T      `json:"node"`
}

type connection[T models.Identity] struct {
	PageInfo pageInfo  `json:"pageInfo"`
	Edges    []edge[T] `json:"edges"`
}

var pageInfoType = graphql.NewObject(graphql.ObjectConfig{
	Name: "PageInfo",
	Fields: graphql.Fields{
		"hasNextPage": &graphql.Field{
			Type: graphql.Boolean,
		},
		"hasPreviousPage": &graphql.Field{
			Type: graphql.Boolean,
		},
		"startCursor": &graphql.Field{
			Type: graphql.String,
		},
		"endCursor": &graphql.Field{
			Type: graphql.String,
		},
	},
})

func toConnection[T models.Identity](first int, results []T) connection[T] {
	var conn connection[T]
	if len(results) == 0 {
		return conn
	}
	for _, r := range results {
		conn.Edges = append(conn.Edges, edge[T]{
			Cursor: r.Identity(),
			Node:   r,
		})
	}

	conn.PageInfo.HasPreviousPage = false
	conn.PageInfo.HasNextPage = len(results) == first
	conn.PageInfo.StartCursor = results[0].Identity()
	conn.PageInfo.EndCursor = results[len(results)-1].Identity()

	return conn
}

func toConnectionType(object *graphql.Object) *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: fmt.Sprintf("%sConnection", object.Name()),
		Fields: graphql.Fields{
			"edges": &graphql.Field{
				Type: graphql.NewList(graphql.NewObject(graphql.ObjectConfig{
					Name: fmt.Sprintf("%sEdge", object.Name()),
					Fields: graphql.Fields{
						"cursor": &graphql.Field{
							Type: graphql.ID,
						},
						"node": &graphql.Field{
							Type: object,
						},
					},
				})),
			},
			"pageInfo": &graphql.Field{
				Type: pageInfoType,
			},
		},
	})
}
