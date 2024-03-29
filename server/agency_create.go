package main

import (
	"database/sql"
	"fmt"
	"log/slog"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"github.com/go-playground/validator/v10"
	"github.com/graphql-go/graphql"
	"github.com/jmoiron/sqlx"
)

// createAgencyMutation returns the `createAgency` mutation field.
//
// `createAgency` allows a pager admin to create a new agency in the system on
// behalf of a real world agency. The agency is created in an `INACTIVE` status
func createAgencyMutation(
	logger *slog.Logger,
	types graphTypes,
	validate *validator.Validate,
	authz *authzed.Client,
	db *sqlx.DB) *graphql.Field {
	// createAgencyInput represents the fields needed to create a new agency.
	type createAgencyInput struct {
		Name string `json:"name"`
	}

	// createAgencyInputType is the graphql input type for the createAgency
	// mutation.
	var createAgencyInputType = graphql.InputObjectConfig{
		Name: "CreateAgencyInput",
		Fields: graphql.InputObjectConfigFieldMap{
			"name": &graphql.InputObjectFieldConfig{
				Type: graphql.String,
			},
		},
	}

	// toCreateAgencyInput converts a `map[string]interface{}` into a
	// `createAgencyInput`.
	toCreateAgencyInput := func(args map[string]interface{}) createAgencyInput {
		var input createAgencyInput
		// Name
		name, ok := args["name"].(string)
		if !ok {
			name = ""
		}
		input.Name = name
		return input
	}

	// createAgencyPayload is the struct representation of the result of a
	// successful CreateAgency mutation.
	type createAgencyPayload struct {
		Agency agency `json:"agency"`
	}

	// createAgencyPayloadType is the graphql representation of the result of a
	// successful CreateAgency mutation.
	var createAgencyPayloadType = graphql.NewObject(graphql.ObjectConfig{
		Name: "CreateAgencyPayload",
		Fields: graphql.Fields{
			"agency": &graphql.Field{
				Type: types.agency,
			},
		},
	})

	return &graphql.Field{
		Name: "createAgency",
		Type: toResultType[createAgencyPayload](
			createAgencyPayloadType,
			baseErrorType,
			validationErrorType,
			authzErrorType),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewInputObject(createAgencyInputType),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			requestContext := p.Context.Value(pagerContextKey{}).(pagerContext)

			authzCheck, err := authz.CheckPermission(p.Context, &v1.CheckPermissionRequest{
				Resource: &v1.ObjectReference{
					ObjectType: "platform",
					ObjectId:   "platform",
				},
				Permission: "create_agency",
				Subject: &v1.SubjectReference{
					Object: &v1.ObjectReference{
						ObjectType: "user",
						ObjectId:   requestContext.User,
					},
				},
			})

			if err != nil {
				return nil, err
			}

			// Validate input model returning a validation error if one occurs
			input := toCreateAgencyInput(p.Args["input"].(map[string]interface{}))
			if err := validate.Struct(input); err != nil {
				return err, nil
			}

			// Check that this user is authorized to create_agency on the platform
			if authzCheck.Permissionship != v1.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION {
				return newAuthzError(requestContext.User, "platform", "create_agency"), nil
			}

			tx, err := db.BeginTxx(p.Context, &sql.TxOptions{
				Isolation: sql.LevelSerializable,
			})
			if err != nil {
				return nil, err
			}

			var agency agency
			if err := tx.QueryRowxContext(
				p.Context,
				"INSERT INTO agency (name, created_by, modified_by) VALUES ($1, $2, $3) RETURNING id, name, created, created_by, modified, modified_by;",
				input.Name,
				requestContext.User,
				requestContext.User,
			).StructScan(&agency); err != nil {
				return nil, err
			}

			_, err = authz.WriteRelationships(p.Context, &v1.WriteRelationshipsRequest{
				Updates: []*v1.RelationshipUpdate{
					{
						Operation: v1.RelationshipUpdate_OPERATION_CREATE,
						Relationship: &v1.Relationship{
							Resource: &v1.ObjectReference{
								ObjectType: "agency",
								ObjectId:   agency.ID,
							},
							Subject: &v1.SubjectReference{
								Object: &v1.ObjectReference{
									ObjectType: "platform",
									ObjectId:   "platform",
								},
							},
							Relation: "platform",
						},
					},
				},
			})

			if err != nil {
				if txerr := tx.Rollback(); txerr != nil {
					return fmt.Errorf("failed to rollback transaction: %w rollback reason: %w", txerr, err), nil
				}

				return nil, err
			}

			if err := tx.Commit(); err != nil {
				return nil, err
			}

			return createAgencyPayload{
				Agency: agency,
			}, nil
		},
	}
}
