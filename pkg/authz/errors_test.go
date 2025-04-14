package authz_test

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/jsmithdenverdev/pager/pkg/authz"
	"github.com/stretchr/testify/assert"
)

func TestUnauthorizedError(t *testing.T) {
	testcases := map[string]struct {
		entityType  string
		entityId    string
		actionType  string
		actionId    string
		expectedErr string
	}{
		"general action": {
			entityType:  "pager::User",
			entityId:    "1234567890",
			actionType:  "pager::Action",
			actionId:    "1234567890",
			expectedErr: "user is not authorized to perform action pager::Action::1234567890 on resource pager::User::1234567890",
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			err := authz.NewUnauthorizedError(
				&types.EntityIdentifier{
					EntityType: aws.String("pager::User"),
					EntityId:   aws.String("1234567890"),
				},
				&types.ActionIdentifier{
					ActionType: aws.String("pager::Action"),
					ActionId:   aws.String("1234567890"),
				},
			)

			assert.Equal(t, tc.expectedErr, err.Error())
		})

	}
}
