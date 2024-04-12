package service

import (
	"context"
	"log/slog"

	"github.com/graph-gophers/dataloader/v7"
	"github.com/jmoiron/sqlx"
	"github.com/jsmithdenverdev/pager/authz"
	"github.com/jsmithdenverdev/pager/models"
)

// PageOrder  represents the sort order in a request to list pages.
type PageOrder int

// PageDeliveryOrder represents the sort order in a request to list page
// deliveries.
type PageDeliveryOrder int

const (
	PageOrderCreatedAsc PageOrder = iota
	PageOrderCreatedDesc
	PageOrderModifiedAsc
	PageOrderModifiedDesc
	PageOrderNameAsc
	PageOrderNameDesc
)

const (
	PageDeliveryOrderCreatedAsc PageDeliveryOrder = iota
	PageDeliveryOrderCreatedDesc
	PageDeliveryOrderModifiedAsc
	PageDeliveryOrderModifiedDesc
	PageDeliveryOrderNameAsc
	PageDeliveryOrderNameDesc
)

var (
	pageOrderNames = [...]string{
		"CREATED_ASC",
		"CREATED_DESC",
		"MODIFIED_ASC",
		"MODIFIED_DESC",
		"NAME_ASC",
		"NAME_DESC",
	}
	pageDeliveryOrderNames = [...]string{
		"CREATED_ASC",
		"CREATED_DESC",
		"MODIFIED_ASC",
		"MODIFIED_DESC",
		"NAME_ASC",
		"NAME_DESC",
	}
	pageOrderMap = map[PageOrder]string{
		PageOrderCreatedAsc:   "created ASC",
		PageOrderCreatedDesc:  "created DESC",
		PageOrderModifiedAsc:  "modified ASC",
		PageOrderModifiedDesc: "modified DESC",
		PageOrderNameAsc:      "name ASC",
		PageOrderNameDesc:     "name DESC",
	}
	pageDeliveryOrderMap = map[PageDeliveryOrder]string{
		PageDeliveryOrderCreatedAsc:   "created ASC",
		PageDeliveryOrderCreatedDesc:  "created DESC",
		PageDeliveryOrderModifiedAsc:  "modified ASC",
		PageDeliveryOrderModifiedDesc: "modified DESC",
		PageDeliveryOrderNameAsc:      "name ASC",
		PageDeliveryOrderNameDesc:     "name DESC",
	}
)

func (order PageDeliveryOrder) String() string {
	return pageDeliveryOrderNames[order]
}

func (order PageOrder) String() string {
	return pageOrderNames[order]
}

type PagePagination struct {
	First  int
	Filter struct {
		AgencyID string
		UserID   string
	}
	After string
	Order PageOrder
}

type PageDeliveryPagination struct {
	First  int
	Filter struct {
		AgencyID string
		UserID   string
	}
	After string
	Order PageDeliveryOrder
}

// DeviceService exposes all operations that can be performed on or for devices.
type PageService struct {
	ctx                      context.Context
	user                     string
	authclient               *authz.Client
	db                       *sqlx.DB
	logger                   *slog.Logger
	listPagesDataloader      *dataloader.Loader[PagePagination, []models.Page]
	readPageDataloader       *dataloader.Loader[string, models.Page]
	listDeliveriesDataloader *dataloader.Loader[PageDeliveryPagination, []models.PageDelivery]
	readDeliveryDataloader   *dataloader.Loader[string, models.PageDelivery]
}

// NewDeviceService creates a new DeviceService. A pointer to the service is
// returned.
func NewPageService(
	ctx context.Context,
	user string,
	authz *authz.Client,
	db *sqlx.DB,
	logger *slog.Logger,
) *PageService {
	return &PageService{
		ctx:        ctx,
		user:       user,
		authclient: authz,
		db:         db,
		logger:     logger,
	}
}

func (service *PageService) CreatePage(agencyId string, content string) (models.Page, error) {
	var page models.Page
	return page, nil
}

func (service *PageService) ReadPage(id string) (models.Page, error) {
	return service.readPageDataloader.Load(service.ctx, id)()
}

func (service *PageService) UpdatePage(content string) (models.Page, error) {
	var page models.Page
	return page, nil
}

func (service *PageService) ListPages(pagination PagePagination) ([]models.Page, error) {
	return service.listPagesDataloader.Load(service.ctx, pagination)()
}

func (service *PageService) DeletePage(id string) error {
	return nil
}

func (service *PageService) DeliverPage(pageId string) error {
	return nil
}

func (service *PageService) ReadDelivery(id string) (models.PageDelivery, error) {
	return service.readDeliveryDataloader.Load(service.ctx, id)()
}

func (service *PageService) ListDeliveries(pagination PageDeliveryPagination) ([]models.PageDelivery, error) {
	return service.listDeliveriesDataloader.Load(service.ctx, pagination)()
}
