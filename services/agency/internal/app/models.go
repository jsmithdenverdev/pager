package app

type agency struct {
	PK     string `dynamodbav:"pk"`
	SK     string `dynamodbav:"sk"`
	Type   string `dynamodbav:"type"`
	Name   string `dynamodbav:"name"`
	Status string `dynamodbav:"status"`
}

type membership struct {
	PK   string `dynamodbav:"pk"`
	SK   string `dynamodbav:"sk"`
	Type string `dynamodbav:"type"`
	Name string `dynamodbav:"name"`
	Role string `dynamodbav:"role"`
}
