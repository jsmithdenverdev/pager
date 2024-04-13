package schema

import (
	"github.com/graphql-go/graphql"
	"github.com/jsmithdenverdev/pager/models"
	"github.com/jsmithdenverdev/pager/service"
)

// createPageInput represents the fields needed to create a new page.
type createPageInput struct {
	AgencyID string `json:"agencyId" validate:"required,uuid"`
	Content  string `json:"content" validate:"min=1"`
	Deliver  bool   `json:"deliver"`
}

// createPageInputType is the graphql input type for the createPage
// mutation.
var createPageInputType = graphql.InputObjectConfig{
	Name: "CreatePageInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"agencyId": &graphql.InputObjectFieldConfig{
			Type: graphql.ID,
		},
		"content": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"deliver": &graphql.InputObjectFieldConfig{
			Type: graphql.Boolean,
		},
	},
}

// toCreatePageInput converts a `map[string]interface{}` into a
// `createPageInput` performing validation on the model and returning any
// errors.
func toCreatePageInput(args map[string]interface{}) (createPageInput, error) {
	var input createPageInput
	content, ok := args["content"].(string)
	if !ok {
		content = ""
	}
	input.Content = content
	agencyId, ok := args["agencyId"].(string)
	if !ok {
		agencyId = ""
	}
	input.AgencyID = agencyId
	deliver, ok := args["deliver"].(bool)
	if !ok {
		deliver = false
	}
	input.Deliver = deliver
	return input, validator.Struct(input)
}

// createPagePayload is the struct representation of the result of a
// successful CreatePage mutation.
type createPagePayload struct {
	Page models.Page `json:"page"`
}

// createPagePayloadType is the graphql representation of the result of a
// successful CreatePage mutation.
var createPagePayloadType = graphql.NewObject(graphql.ObjectConfig{
	Name: "CreatePagePayload",
	Fields: graphql.Fields{
		"page": &graphql.Field{
			Type: pageType,
		},
	},
})

// createPageMutation is the field definition for the createPage mutation.
var createPageMutation = &graphql.Field{
	Name: "createPage",
	Type: createPagePayloadType,
	Args: graphql.FieldConfigArgument{
		"input": &graphql.ArgumentConfig{
			Type: graphql.NewInputObject(createPageInputType),
		},
	},
	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
		return func() (interface{}, error) {
			var payload createPagePayload
			input, err := toCreatePageInput(p.Args["input"].(map[string]interface{}))
			if err != nil {
				return payload, err
			}
			svc := p.Context.Value(service.ContextKeyPageService).(*service.PageService)
			page, err := svc.CreatePage(input.AgencyID, input.Content, input.Deliver)
			if err != nil {
				return payload, err
			}
			payload.Page = page
			return payload, nil
		}, nil
	},
}
