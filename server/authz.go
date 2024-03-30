package main

import (
	"context"
	"io"
	"log/slog"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"github.com/graph-gophers/dataloader/v7"
)

// newCheckPermissionDataLoader returns a request scoped data loader used to
// check permissions through authzed.
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

// newLookupResourcesDataloader returns a request scoped data loader used to
// lookup the resources a given subject has access to through authzed.
func newLookupResourcesDataloader(logger *slog.Logger, authz *authzed.Client) *dataloader.Loader[*v1.LookupResourcesRequest, []*v1.LookupResourcesResponse] {
	return dataloader.NewBatchedLoader[*v1.LookupResourcesRequest, []*v1.LookupResourcesResponse](
		func(ctx context.Context, requests []*v1.LookupResourcesRequest) []*dataloader.Result[[]*v1.LookupResourcesResponse] {
			var results []*dataloader.Result[[]*v1.LookupResourcesResponse]
			for _, request := range requests {
				r, err := authz.LookupResources(ctx, request)
				if err != nil {
					results = append(results, &dataloader.Result[[]*v1.LookupResourcesResponse]{
						Error: err,
					})
				}

				var resources []*v1.LookupResourcesResponse

				for {
					resource, err := r.Recv()
					if err != nil {
						if err == io.EOF {
							results = append(results, &dataloader.Result[[]*v1.LookupResourcesResponse]{
								Data: resources,
							})
						} else {
							results = append(results, &dataloader.Result[[]*v1.LookupResourcesResponse]{
								Error: err,
							})
						}
						break
					}
					resources = append(resources, resource)
				}
			}
			return results
		})
}
