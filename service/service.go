package service

type (
	// A Service represent a long-lived application
	Service struct {
		Name        string
		Directory   string
		Command     []string
		Environment map[string]string
	}

	// A Manager manages services
	Manager interface {
		Install(service Service) error
		Uninstall(serviceName string) error
		List() ([]string, error)
	}
)
