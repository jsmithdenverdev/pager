package schema

import (
	"github.com/graphql-go/graphql"
	"github.com/jsmithdenverdev/pager/models"
	"github.com/jsmithdenverdev/pager/service"
)

// deliverPageInput represents the fields needed to create a new page.
type deliverPageInput struct {
	AgencyID string `json:"agencyId" validate:"required,uuid"`
	PageID   string `json:"content" validate:"min=1"`
}

// deliverPageInputType is the graphql input type for the deliverPage
// mutation.
var deliverPageInputType = graphql.InputObjectConfig{
	Name: "DeliverPageInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"agencyId": &graphql.InputObjectFieldConfig{
			Type: graphql.ID,
		},
		"pageId": &graphql.InputObjectFieldConfig{
			Type: graphql.ID,
		},
	},
}

// toDeliverPageInput converts a `map[string]interface{}` into a
// `deliverPageInput` performing validation on the model and returning any
// errors.
func toDeliverPageInput(args map[string]interface{}) (deliverPageInput, error) {
	var input deliverPageInput
	agencyId, ok := args["agencyId"].(string)
	if !ok {
		agencyId = ""
	}
	input.AgencyID = agencyId
	pageId, ok := args["pageId"].(string)
	if !ok {
		pageId = ""
	}
	input.PageID = pageId
	return input, validator.Struct(input)
}

// deliverPagePayload is the struct representation of the result of a
// successful DeliverPage mutation.
type deliverPagePayload struct {
	Page models.Page `json:"page"`
}

// deliverPagePayloadType is the graphql representation of the result of a
// successful DeliverPage mutation.
var deliverPagePayloadType = graphql.NewObject(graphql.ObjectConfig{
	Name: "DeliverPagePayload",
	Fields: graphql.Fields{
		"page": &graphql.Field{
			Type: pageType,
		},
	},
})

// deliverPageMutation is the field definition for the deliverPage mutation.
var deliverPageMutation = &graphql.Field{
	Name: "deliverPage",
	Type: deliverPagePayloadType,
	Args: graphql.FieldConfigArgument{
		"input": &graphql.ArgumentConfig{
			Type: graphql.NewInputObject(deliverPageInputType),
		},
	},
	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		return func() (interface{}, error) {
			var payload deliverPagePayload
			input, err := toDeliverPageInput(p.Args["input"].(map[string]interface{}))
			if err != nil {
				return payload, err
			}
			svc := p.Context.Value(service.ContextKeyPageService).(*service.PageService)
			page, err := svc.DeliverPage(input.AgencyID, input.PageID)
			if err != nil {
				return payload, err
			}
			payload.Page = page
			return payload, nil
		}, nil
	},
}
