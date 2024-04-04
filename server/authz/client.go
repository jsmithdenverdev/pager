package authz

import (
	"context"
	"io"
	"log/slog"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"github.com/graph-gophers/dataloader/v7"
)

func bulkCheckPermissionsDataloader(authz *authzed.Client) *dataloader.Loader[*v1.CheckBulkPermissionsRequest, *v1.CheckBulkPermissionsResponse] {
	return dataloader.NewBatchedLoader(func(ctx context.Context, requests []*v1.CheckBulkPermissionsRequest) []*dataloader.Result[*v1.CheckBulkPermissionsResponse] {
		results := make([]*dataloader.Result[*v1.CheckBulkPermissionsResponse], len(requests))
		for i, request := range requests {
			result, err := authz.CheckBulkPermissions(ctx, request)
			results[i] = &dataloader.Result[*v1.CheckBulkPermissionsResponse]{
				Data:  result,
				Error: err,
			}
		}
		return results
	})
}

func listDataloader(authz *authzed.Client) *dataloader.Loader[*v1.LookupResourcesRequest, []*v1.LookupResourcesResponse] {
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

type Client struct {
	ctx                            context.Context
	userId                         string
	authzed                        *authzed.Client
	logger                         *slog.Logger
	bulkCheckPermissionsDataloader *dataloader.Loader[*v1.CheckBulkPermissionsRequest, *v1.CheckBulkPermissionsResponse]
	listDataloader                 *dataloader.Loader[*v1.LookupResourcesRequest, []*v1.LookupResourcesResponse]
}

func NewClient(ctx context.Context, authzed *authzed.Client, logger *slog.Logger, userId string) *Client {
	return &Client{
		ctx:                            ctx,
		userId:                         userId,
		authzed:                        authzed,
		bulkCheckPermissionsDataloader: bulkCheckPermissionsDataloader(authzed),
		listDataloader:                 listDataloader(authzed),
	}
}

func (client *Client) Authorize(permission permission, resource Resource) (bool, error) {
	result, err := client.bulkCheckPermissionsDataloader.Load(client.ctx, &v1.CheckBulkPermissionsRequest{
		Items: []*v1.CheckBulkPermissionsRequestItem{
			{
				Resource: &v1.ObjectReference{
					ObjectType: resource.Type,
					ObjectId:   resource.ID,
				},
				Permission: string(permission),
				Subject: &v1.SubjectReference{
					Object: &v1.ObjectReference{
						ObjectType: "user",
						ObjectId:   client.userId,
					},
				},
			},
		},
	})()

	if err != nil {
		client.logger.ErrorContext(client.ctx, "error: CheckPermission failed", "error", err)
		return false, err
	}

	return result.Pairs[0].GetItem().Permissionship == v1.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION, nil
}

func (client *Client) BatchAuthorize(permission permission, resources []Resource) ([]bool, error) {
	items := make([]*v1.CheckBulkPermissionsRequestItem, len(resources))
	for i, resource := range resources {
		items[i] = &v1.CheckBulkPermissionsRequestItem{
			Resource: &v1.ObjectReference{
				ObjectType: resource.Type,
				ObjectId:   resource.ID,
			},
			Permission: string(permission),
			Subject: &v1.SubjectReference{
				Object: &v1.ObjectReference{
					ObjectType: "user",
					ObjectId:   client.userId,
				},
			},
		}
	}

	permissions := make([]bool, len(resources))

	results, err := client.bulkCheckPermissionsDataloader.Load(
		client.ctx,
		&v1.CheckBulkPermissionsRequest{
			Items: items,
		})()

	if err != nil {
		return permissions, err
	}

	for i, result := range results.Pairs {
		permissions[i] = result.GetItem().Permissionship == v1.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION
	}

	return permissions, nil
}

func (client *Client) List(permission permission, resource Resource) ([]string, error) {
	var resourceIds []string
	resources, err := client.listDataloader.Load(client.ctx, &v1.LookupResourcesRequest{
		ResourceObjectType: resource.Type,
		Subject: &v1.SubjectReference{
			Object: &v1.ObjectReference{
				ObjectType: "user",
				ObjectId:   client.userId,
			},
		},
		Permission: string(permission),
	})()

	if err != nil {
		return resourceIds, err
	}

	for _, resource := range resources {
		resourceIds = append(resourceIds, resource.ResourceObjectId)
	}

	return resourceIds, nil
}

func (client *Client) WritePermission(relationship string, resource Resource, subject Resource) error {
	_, err := client.authzed.WriteRelationships(client.ctx, &v1.WriteRelationshipsRequest{
		Updates: []*v1.RelationshipUpdate{
			{
				Operation: v1.RelationshipUpdate_OPERATION_CREATE,
				Relationship: &v1.Relationship{
					Resource: &v1.ObjectReference{
						ObjectType: resource.Type,
						ObjectId:   resource.ID,
					},
					Subject: &v1.SubjectReference{
						Object: &v1.ObjectReference{
							ObjectType: subject.Type,
							ObjectId:   subject.ID,
						},
					},
					Relation: relationship,
				},
			},
		},
	})

	return err
}
