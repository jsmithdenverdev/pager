package service

type contextKey string

const (
	ContextKeyAgencyService contextKey = "agency_service"
	ContextKeyUserService   contextKey = "user_service"
	ContextKeyDeviceService contextKey = "device_service"
)
