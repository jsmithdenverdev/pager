package app

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func addRoutes(mux *http.ServeMux, config Config, logger *slog.Logger, client *dynamodb.Client) {
	mux.Handle(fmt.Sprintf("GET /%s", config.Environment), listMyMemberships(config, logger, client))
	mux.Handle(fmt.Sprintf("GET /%s/{id}", config.Environment), readAgency(config, logger, client))
	mux.Handle(fmt.Sprintf("GET /%s/{id}/memberships", config.Environment), listAgencyMemberships(config, logger, client))
}
