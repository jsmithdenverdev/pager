package authz

import (
	"context"
	"io"
	"log/slog"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"github.com/graph-gophers/dataloader/v7"
)

func newCheckPermissionDataloader(authz *authzed.Client) *dataloader.Loader[*v1.CheckPermissionRequest, *v1.CheckPermissionResponse] {
	return dataloader.NewBatchedLoader[*v1.CheckPermissionRequest, *v1.CheckPermissionResponse](
		func(ctx context.Context, requests []*v1.CheckPermissionRequest) []*dataloader.Result[*v1.CheckPermissionResponse] {
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

func newListResourcesDataLoader(authz *authzed.Client) *dataloader.Loader[*v1.LookupResourcesRequest, []*v1.LookupResourcesResponse] {
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

type Client interface {
	CheckPermission(permission, resourceType, resourceId string) (bool, error)
	List(permission, resourceType string) ([]string, error)
}

type client struct {
	ctx                       context.Context
	userId                    string
	authzed                   *authzed.Client
	logger                    *slog.Logger
	checkPermissionDataloader *dataloader.Loader[*v1.CheckPermissionRequest, *v1.CheckPermissionResponse]
	listResourcesDataLoader   *dataloader.Loader[*v1.LookupResourcesRequest, []*v1.LookupResourcesResponse]
}

func NewClient(ctx context.Context, authzed *authzed.Client, logger *slog.Logger, userId string) Client {
	return &client{
		ctx:                       ctx,
		userId:                    userId,
		authzed:                   authzed,
		checkPermissionDataloader: newCheckPermissionDataloader(authzed),
		listResourcesDataLoader:   newListResourcesDataLoader(authzed),
	}
}

func (client *client) CheckPermission(permission, resourceType, resourceId string) (bool, error) {
	result, err := client.checkPermissionDataloader.Load(
		client.ctx,
		&v1.CheckPermissionRequest{
			Resource: &v1.ObjectReference{
				ObjectType: resourceType,
				ObjectId:   resourceId,
			},
			Permission: permission,
			Subject: &v1.SubjectReference{
				Object: &v1.ObjectReference{
					ObjectType: "user",
					ObjectId:   client.userId,
				},
			},
		})()

	if err != nil {
		client.logger.ErrorContext(client.ctx, "error: CheckPermission failed", "error", err)
		return false, err
	}

	return result.Permissionship == v1.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION, nil
}

func (client *client) List(permission, resourceType string) ([]string, error) {
	var resourceIds []string
	resources, err := client.listResourcesDataLoader.Load(client.ctx, &v1.LookupResourcesRequest{
		ResourceObjectType: resourceType,
		Subject: &v1.SubjectReference{
			Object: &v1.ObjectReference{
				ObjectType: "user",
				ObjectId:   client.userId,
			},
		},
		Permission: permission,
	})()

	if err != nil {
		return resourceIds, err
	}

	for _, resource := range resources {
		resourceIds = append(resourceIds, resource.ResourceObjectId)
	}

	return resourceIds, nil
}
