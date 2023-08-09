package main

type systemConnections struct{}
type systemRepositories struct{}
type systemServices struct{}
type ServiceProvider struct {
	services systemServices
}

func (s *ServiceProvider) initiate(config ApiConfig) {
	connections := s.setUpConnections()
	repositories := s.registerRepositories(connections)
	s.registerServices(repositories)
}

func (s *ServiceProvider) setUpConnections() systemConnections {
	return systemConnections{}
}

func (s *ServiceProvider) registerRepositories(connections systemConnections) systemRepositories {
	return systemRepositories{}
}
func (s *ServiceProvider) registerServices(repositories systemRepositories) {
	s.services = systemServices{}
}

func (s *ServiceProvider) cleanUp() {
	// close connections
}
