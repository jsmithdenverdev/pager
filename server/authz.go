package main

import (
	"context"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"github.com/graph-gophers/dataloader/v7"
)

func newCheckPermissionDataLoader(authz *authzed.Client) *dataloader.Loader[*v1.CheckPermissionRequest, *v1.CheckPermissionResponse] {
	return dataloader.NewBatchedLoader(
		func(ctx context.Context,
			requests []*v1.CheckPermissionRequest) []*dataloader.Result[*v1.CheckPermissionResponse] {
			var results []*dataloader.Result[*v1.CheckPermissionResponse]
			for _, request := range requests {
				result, err := authz.CheckPermission(ctx, request)
				results = append(results, &dataloader.Result[*v1.CheckPermissionResponse]{
					Data:  result,
					Error: err,
				})
			}
			return results
		})
}
