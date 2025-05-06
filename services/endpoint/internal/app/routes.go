package app

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func addRoutes(mux *http.ServeMux, config Config, logger *slog.Logger, client *dynamodb.Client) {
	mux.Handle(fmt.Sprintf("GET /%s", config.Environment), listEndpoints(config, logger, client))
	// mux.Handle(fmt.Sprintf("GET /%s/registrations/agencies/{agencyid}", config.Environment), listRegistrations(config, logger, client))
}
