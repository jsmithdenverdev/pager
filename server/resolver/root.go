package resolver

type Root struct{}

func (r *Root) Query() *Query {
	return &Query{}
}

func (r *Root) Mutation() *Mutation {
	return &Mutation{}
}
