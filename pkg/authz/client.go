package authz

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions"
	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions/types"
	"github.com/aws/aws-sdk-go/aws"
)

type optionFunc func(*Client)

type Client struct {
	userInfo      User
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

	// Create entity definitions to hold attributes for the entities supplied in
	// an authz request. The default set of entitiy definitions are for a user
	// and include the users entitlements.
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
				},
			},
		},
	}

	// If the user is a member of agencies add those agencies to the auth request
	// context.
	if len(c.userInfo.Agencies) > 0 {
		// Encode the users agencies into a slice of entity identifier attribute value
		var agencyAttributeValues []types.AttributeValue
		for agency := range c.userInfo.Agencies {
			agencyAttributeValues = append(agencyAttributeValues, &types.AttributeValueMemberEntityIdentifier{
				Value: types.EntityIdentifier{
					EntityType: aws.String("pager::Agency"),
					EntityId:   aws.String(agency),
				},
			})
		}
		// entityDefinitions.Value[0] is the user
		entityDefinitions.Value[0].Attributes["agencies"] = &types.AttributeValueMemberSet{
			Value: agencyAttributeValues,
		}
	}

	// If the user is making a request for a specific agency add the agency to
	// the auth request context.
	if len(c.userInfo.ActiveAgency) > 0 {
		// entityDefinitions.Value[0] is the user
		entityDefinitions.Value[0].Attributes["currentAgency"] = &types.AttributeValueMemberEntityIdentifier{
			Value: types.EntityIdentifier{
				EntityType: aws.String("pager::Agency"),
				EntityId:   aws.String(c.userInfo.ActiveAgency),
			},
		}
		if agency, ok := c.userInfo.Agencies[c.userInfo.ActiveAgency]; ok {
			entityDefinitions.Value = append(
				entityDefinitions.Value,
				types.EntityItem{
					Identifier: &types.EntityIdentifier{
						EntityType: aws.String("pager::Agency"),
						EntityId:   aws.String(c.userInfo.ActiveAgency),
					},
					Attributes: map[string]types.AttributeValue{
						"membership": &types.AttributeValueMemberEntityIdentifier{
							Value: types.EntityIdentifier{
								EntityType: aws.String("pager::Membership"),
								EntityId:   aws.String(agency.Role),
							},
						},
					},
				})
		}

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

func WithUserInfo(userInfo User) optionFunc {
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
