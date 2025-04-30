package app

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func addRoutes(mux *http.ServeMux, config Config, logger *slog.Logger, client *dynamodb.Client) {
	mux.Handle(fmt.Sprintf("GET /%s", config.Environment), listAgencies(config, logger, client))
	mux.Handle(fmt.Sprintf("GET /%s/{id}", config.Environment), readAgency(config, logger, client))
	mux.Handle(fmt.Sprintf("GET /%s/{id}/members", config.Environment), listMemberships(config, logger, client))

	mux.Handle(fmt.Sprintf("POST /%s", config.Environment), createAgency(config, logger, client))
	mux.Handle(fmt.Sprintf("POST /%s/{id}/members", config.Environment), createMembership(config, logger, client))
}
