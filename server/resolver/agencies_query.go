package resolver

import (
	"context"
	"fmt"

	"github.com/jsmithdenverdev/pager/service"
)

// AgenciesOrder represents the sort order in a request to list agencies.
type AgenciesOrder int32

const (
	AgenciesOrderCreatedAsc AgenciesOrder = iota
	AgenciesOrderCreatedDesc
	AgenciesOrderModifiedAsc
	AgenciesOrderModifiedDesc
	AgenciesOrderNameAsc
	AgenciesOrderNameDesc
)

var (
	agenciesOrderNames = [...]string{
		"CREATED_ASC",
		"CREATED_DESC",
		"MODIFIED_ASC",
		"MODIFIED_DESC",
		"NAME_ASC",
		"NAME_DESC",
	}
)

func (order AgenciesOrder) String() string {
	return agenciesOrderNames[order]
}

func (order *AgenciesOrder) Deserialize(str string) {
	for i, o := range agenciesOrderNames {
		if o == str {
			(*order) = AgenciesOrder(i)
			return
		}
	}
	panic("invalid value for enum AgenciesOrder: " + str)
}

func (order *AgenciesOrder) ImplementsGraphQLType(name string) bool {
	return name == "AgenciesOrder"
}

func (order *AgenciesOrder) UnmarshalGraphQL(input interface{}) error {
	var err error
	switch input := input.(type) {
	case string:
		order.Deserialize(input)
	default:
		err = fmt.Errorf("wrong type for AgenciesOrder: %T", input)
	}
	return err
}

func (query *Query) Agencies(ctx context.Context, args struct {
	First int32
	After *string
	Order AgenciesOrder
}) (*[]*Agency, error) {
	svc := ctx.Value(service.ContextKeyAgencyService).(*service.AgencyService)
	var after string
	if args.After != nil {
		after = *args.After
	}
	agencies, err := svc.List(service.AgenciesPagination{
		First: int(args.First),
		After: after,
		Order: service.AgenciesOrder(args.Order),
	})
	if err != nil {
		return nil, err
	}
	var results []*Agency
	for _, agency := range agencies {
		results = append(results, &Agency{agency})
	}
	return &results, nil
}
