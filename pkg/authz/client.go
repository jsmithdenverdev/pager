package authz

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions"
	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions/types"
	"github.com/aws/aws-sdk-go/aws"
)

type optionFunc func(*Client)

type Client struct {
	userInfo      UserInfo
	policyStoreId string
	*verifiedpermissions.Client
}

func NewClient(options ...optionFunc) *Client {
	c := new(Client)
	for _, o := range options {
		o(c)
	}
	return c
}

type IsAuthorizedInput struct {
	Resource *types.EntityIdentifier
	Action   *types.ActionIdentifier
	Entities []types.EntityItem
}

func (c *Client) IsAuthorized(ctx context.Context, input IsAuthorizedInput) (bool, error) {
	// Encode the users entitlements into a slice of string attribute values
	var entitlementAttributeValues []types.AttributeValue
	for _, entitlement := range c.userInfo.Entitlements {
		entitlementAttributeValues = append(entitlementAttributeValues, &types.AttributeValueMemberString{
			Value: string(entitlement),
		})
	}

	// Encode the users accounts into a slice of entity identifier attribute value
	var accountAttributeValues []types.AttributeValue
	for account := range c.userInfo.Accounts {
		accountAttributeValues = append(accountAttributeValues, &types.AttributeValueMemberEntityIdentifier{
			Value: types.EntityIdentifier{
				EntityType: aws.String("pager::Agency"),
				EntityId:   aws.String(account),
			},
		})
	}

	// Create entity definitions to hold attributes for the entities supplied in
	// an authz request. The default set of entitiy definitions are for a user
	// and include the users entitlements, the accounts they are a member of, and
	// optionally the current account.
	entityDefinitions := &types.EntitiesDefinitionMemberEntityList{
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
					"accounts": &types.AttributeValueMemberSet{
						Value: accountAttributeValues,
					},
					"currentAccount": &types.AttributeValueMemberEntityIdentifier{
						Value: types.EntityIdentifier{
							EntityType: aws.String("pager::Account"),
							EntityId:   aws.String(c.userInfo.ActiveAgency),
						},
					},
				},
			},
		},
	}

	// If we're operating in the context of an agency, we need to add an entity
	// definition representing that agency to the request. The agency definition
	// will include the role of the active user within the agency.
	if len(c.userInfo.ActiveAgency) > 0 {
		entityDefinitions.Value = append(
			entityDefinitions.Value,
			types.EntityItem{
				Identifier: &types.EntityIdentifier{
					EntityType: aws.String("pager::Agency"),
					EntityId:   aws.String(c.userInfo.ActiveAgency),
				},
				Attributes: map[string]types.AttributeValue{
					"group": &types.AttributeValueMemberEntityIdentifier{
						Value: types.EntityIdentifier{
							EntityType: aws.String("pager::Group"),
							EntityId:   aws.String(c.userInfo.Accounts[c.userInfo.ActiveAgency].Role),
						},
					},
				},
			})
	}

	entityDefinitions.Value = append(entityDefinitions.Value, input.Entities...)

	// Create a verified permissions request
	authzRequest := verifiedpermissions.IsAuthorizedInput{
		PolicyStoreId: aws.String(c.policyStoreId),
		Principal: &types.EntityIdentifier{
			EntityType: aws.String("pager::User"),
			EntityId:   aws.String(c.userInfo.IPDID),
		},
		Resource: input.Resource,
		Action:   input.Action,
		Entities: entityDefinitions,
	}

	result, err := c.Client.IsAuthorized(ctx, &authzRequest)
	if err != nil {
		return false, err
	}
	return result.Decision == types.DecisionAllow, nil
}

func WithVerifiedPermissionsClient(vpc *verifiedpermissions.Client) optionFunc {
	return func(c *Client) {
		c.Client = vpc
	}
}

func WithUserInfo(userInfo UserInfo) optionFunc {
	return func(c *Client) {
		c.userInfo = userInfo
		return
	}
}

func WithPolicyStoreID(policyStoreId string) optionFunc {
	return func(c *Client) {
		c.policyStoreId = policyStoreId
	}
}
