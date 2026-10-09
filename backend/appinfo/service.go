package appinfo

import "github.com/wailsapp/wails/v3/pkg/application"

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) ServiceName() string {
	return "AppInfoService"
}

func (s *Service) ServiceStartup(ctx application.Context) error {
	return nil
}

func (s *Service) ServiceShutdown() {}

// Version returns the application version.
func (s *Service) Version() string {
	return Version
}
