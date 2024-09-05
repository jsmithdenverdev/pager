package authz

import (
	"context"
	"errors"
	"io"
	"log/slog"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"github.com/graph-gophers/dataloader/v7"
)

func bulkCheckPermissionsDataloader(authz *authzed.Client) *dataloader.Loader[*v1.CheckBulkPermissionsRequest, *v1.CheckBulkPermissionsResponse] {
	return dataloader.NewBatchedLoader(
		func(ctx context.Context, requests []*v1.CheckBulkPermissionsRequest) []*dataloader.Result[*v1.CheckBulkPermissionsResponse] {
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
	return dataloader.NewBatchedLoader(
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
		logger:                         logger,
		bulkCheckPermissionsDataloader: bulkCheckPermissionsDataloader(authzed),
		listDataloader:                 listDataloader(authzed),
	}
}

func (client *Client) Authorize(permission permission, resource Resource) Result {
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
		return Result{
			Authorized: false,
			Error:      err,
		}
	}

	if result.Pairs[0].GetError() != nil {
		return Result{
			Authorized: false,
			Error:      err,
		}
	}

	return Result{
		Authorized: result.Pairs[0].GetItem().Permissionship == v1.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION,
		Error:      nil,
	}
}

func (client *Client) BatchAuthorize(permission permission, resources []Resource) ([]Result, error) {
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

	permissions := make([]Result, len(resources))

	results, err := client.bulkCheckPermissionsDataloader.Load(
		client.ctx,
		&v1.CheckBulkPermissionsRequest{
			Items: items,
		})()

	if err != nil {
		return permissions, err
	}

	for i, result := range results.Pairs {
		// Permission checks from SpiceDB return a result type that can either be
		// the permission or an error. If we have an error we'll fail the entire
		// batch call and return it. This is a bit misleading, because the error may
		// not apply to every item in the set, but I can't think of a better way to
		// handle this for now.
		if err := result.GetError(); err != nil {
			permissions[i] = Result{
				Error:      errors.New(err.Message),
				Authorized: false,
			}
		}
		permissions[i] = Result{
			Error:      nil,
			Authorized: result.GetItem().Permissionship == v1.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION,
		}
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

func (client *Client) WritePermissions(permissions []Permission) error {
	var updates []*v1.RelationshipUpdate
	for _, permission := range permissions {
		updates = append(updates, &v1.RelationshipUpdate{
			Operation: v1.RelationshipUpdate_OPERATION_CREATE,
			Relationship: &v1.Relationship{
				Resource: &v1.ObjectReference{
					ObjectType: permission.Resource.Type,
					ObjectId:   permission.Resource.ID,
				},
				Subject: &v1.SubjectReference{
					Object: &v1.ObjectReference{
						ObjectType: permission.Subject.Type,
						ObjectId:   permission.Subject.ID,
					},
				},
				Relation: permission.Relationship,
			},
		})
	}

	_, err := client.authzed.WriteRelationships(client.ctx, &v1.WriteRelationshipsRequest{
		Updates: updates,
	})

	return err
}
