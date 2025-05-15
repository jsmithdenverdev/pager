package app

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

func addRoutes(mux *http.ServeMux, config Config, logger *slog.Logger, dynamoClient *dynamodb.Client, snsClient *sns.Client) {
	mux.Handle(fmt.Sprintf("GET /%s", config.Environment), listAgencies(config, logger, dynamoClient))
	mux.Handle(fmt.Sprintf("GET /%s/{id}", config.Environment), readAgency(config, logger, dynamoClient))
	mux.Handle(fmt.Sprintf("GET /%s/{id}/members", config.Environment), listMemberships(config, logger, dynamoClient))

	mux.Handle(fmt.Sprintf("POST /%s", config.Environment), createAgency(config, logger, dynamoClient))
	mux.Handle(fmt.Sprintf("POST /%s/{id}/invites", config.Environment), createInvite(config, logger, dynamoClient, snsClient))
	mux.Handle(fmt.Sprintf("POST /%s/{id}/endpoints", config.Environment), registerEndpoint(config, logger, dynamoClient, snsClient))
}
