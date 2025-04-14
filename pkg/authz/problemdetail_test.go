package authz_test

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/jsmithdenverdev/pager/pkg/authz"
	"github.com/stretchr/testify/assert"
)

func TestNewProblemDetail(t *testing.T) {
	pd := authz.NewProblemDetail(authz.UnauthorizedError{
		Entity: &types.EntityIdentifier{
			EntityType: aws.String("pager::User"),
			EntityId:   aws.String("1234567890"),
		},
		Action: &types.ActionIdentifier{
			ActionType: aws.String("pager::Action"),
			ActionId:   aws.String("1234567890"),
		},
	})

	assert.Equal(t, "authorization", pd.Kind())
	assert.Equal(t, "problem detail: authorization", pd.Error())
}
