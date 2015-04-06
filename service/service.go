package service

type (
	Service struct {
		Name        string
		Directory   string
		Command     []string
		Environment map[string]string
	}

	ServiceManager interface {
		Install(service Service) error
		Uninstall(serviceName string) error
		List() ([]string, error)
	}
)
