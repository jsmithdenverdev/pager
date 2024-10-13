package authz

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions"
	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions/types"
	"github.com/aws/aws-sdk-go/aws"
)

type optionFunc func(*client)

type Resource struct {
	Type string
	ID   string
}

type Action struct {
	Type string
	ID   string
}

type Authorizer interface {
	IsAuthorized(context.Context, Resource, Action) (bool, error)
}

type client struct {
	userInfo      UserInfo
	policyStoreId string
	*verifiedpermissions.Client
}

func New(options ...optionFunc) Authorizer {
	c := &client{}
	for _, o := range options {
		o(c)
	}
	return c
}

func (c *client) IsAuthorized(ctx context.Context, resource Resource, action Action) (bool, error) {
	var entitlementAttributeValues []types.AttributeValue
	for _, entitlement := range c.userInfo.Entitlements {
		entitlementAttributeValues = append(entitlementAttributeValues, &types.AttributeValueMemberString{
			Value: string(entitlement),
		})
	}
	authzRequest := verifiedpermissions.IsAuthorizedInput{
		PolicyStoreId: aws.String(c.policyStoreId),
		Principal: &types.EntityIdentifier{
			EntityType: aws.String("pager::User"),
			EntityId:   aws.String(c.userInfo.IPDID),
		},
		Resource: &types.EntityIdentifier{
			EntityType: aws.String(resource.Type),
			EntityId:   aws.String(resource.ID),
		},
		Action: &types.ActionIdentifier{
			ActionType: aws.String(action.Type),
			ActionId:   aws.String(action.ID),
		},
		Entities: &types.EntitiesDefinitionMemberEntityList{
			Value: []types.EntityItem{
				{
					Identifier: &types.EntityIdentifier{
						EntityType: aws.String("pager::User"),
						EntityId:   aws.String(c.userInfo.IPDID),
					},
					Attributes: map[string]types.AttributeValue{
						"entitlements": &types.AttributeValueMemberSet{
							Value: entitlementAttributeValues,
						},
					},
				},
			},
		},
	}

	result, err := c.Client.IsAuthorized(ctx, &authzRequest)
	if err != nil {
		return false, err
	}
	return result.Decision == types.DecisionAllow, nil
}

func WithVerifiedPermissionsClient(vpc *verifiedpermissions.Client) optionFunc {
	return func(c *client) {
		c.Client = vpc
	}
}

func WithUserInfo(userInfo UserInfo) optionFunc {
	return func(c *client) {
		c.userInfo = userInfo
		return
	}
}

func WithPolicyStoreID(policyStoreId string) optionFunc {
	return func(c *client) {
		c.policyStoreId = policyStoreId
	}
}
