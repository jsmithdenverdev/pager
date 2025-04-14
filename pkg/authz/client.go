package authz

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions"
	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions/types"
	"github.com/aws/aws-sdk-go/aws"
)

// optionFunc is a function type that modifies a Client.
type optionFunc func(*Client)

// Client represents a client for interacting with the authorization service.
// It holds user information, a policy store ID, and an AWS Verified Permissions client.
type Client struct {
	user          User
	policyStoreId string
	*verifiedpermissions.Client
}

// NewClient creates a new Client instance, applying any provided options.
func NewClient(options ...optionFunc) *Client {
	c := new(Client)
	for _, o := range options {
		o(c)
	}
	return c
}

// IsAuthorizedInput represents the input required to check if an action is authorized.
type IsAuthorizedInput struct {
	Resource *types.EntityIdentifier
	Action   *types.ActionIdentifier
	Entities []types.EntityItem
}

// IsAuthorized checks if the specified action is authorized for the given entities.
// It encodes user entitlements and creates entity definitions for the authorization request.
func (c *Client) IsAuthorized(ctx context.Context, input IsAuthorizedInput) (bool, error) {
	// Encode the users entitlements into a slice of string attribute values
	var entitlementAttributeValues []types.AttributeValue
	for _, entitlement := range c.user.Entitlements {
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
					EntityId:   aws.String(c.user.IPDID),
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
	if len(c.user.Agencies) > 0 {
		// Encode the users agencies into a slice of entity identifier attribute value
		var agencyAttributeValues []types.AttributeValue
		for agency := range c.user.Agencies {
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
	if len(c.user.ActiveAgency) > 0 {
		// entityDefinitions.Value[0] is the user
		entityDefinitions.Value[0].Attributes["currentAgency"] = &types.AttributeValueMemberEntityIdentifier{
			Value: types.EntityIdentifier{
				EntityType: aws.String("pager::Agency"),
				EntityId:   aws.String(c.user.ActiveAgency),
			},
		}
		if agency, ok := c.user.Agencies[c.user.ActiveAgency]; ok {
			entityDefinitions.Value = append(
				entityDefinitions.Value,
				types.EntityItem{
					Identifier: &types.EntityIdentifier{
						EntityType: aws.String("pager::Agency"),
						EntityId:   aws.String(c.user.ActiveAgency),
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
			EntityId:   aws.String(c.user.IPDID),
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

// WithVerifiedPermissionsClient returns an option that sets the Verified Permissions client.
func WithVerifiedPermissionsClient(vpc *verifiedpermissions.Client) optionFunc {
	return func(c *Client) {
		c.Client = vpc
	}
}

// WithUserInfo returns an option that sets the user information.
func WithUserInfo(userInfo User) optionFunc {
	return func(c *Client) {
		c.user = userInfo
	}
}

// WithPolicyStoreID returns an option that sets the policy store ID.
func WithPolicyStoreID(policyStoreId string) optionFunc {
	return func(c *Client) {
		c.policyStoreId = policyStoreId
	}
}
