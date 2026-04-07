package multireconciler

import "sync"

// Store holds all resource types.
// TODO: Implement multi-type storage with owner-based lookups.
type Store struct {
	mu           sync.RWMutex
	applications map[string]*Application
	services     map[string]*Service
	endpoints    map[string]*Endpoint
}

func NewStore() *Store {
	return &Store{
		applications: make(map[string]*Application),
		services:     make(map[string]*Service),
		endpoints:    make(map[string]*Endpoint),
	}
}

// Application CRUD
func (s *Store) GetApplication(name string) (*Application, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.applications[name]
	return a, ok
}

func (s *Store) PutApplication(a *Application) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.applications[a.Metadata.Name] = a
}

func (s *Store) DeleteApplication(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.applications, name)
}

func (s *Store) ListApplications() []*Application {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*Application, 0, len(s.applications))
	for _, a := range s.applications {
		result = append(result, a)
	}
	return result
}

// Service CRUD
func (s *Store) GetService(name string) (*Service, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	svc, ok := s.services[name]
	return svc, ok
}

func (s *Store) PutService(svc *Service) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.services[svc.Metadata.Name] = svc
}

func (s *Store) DeleteService(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.services, name)
}

func (s *Store) ListServices() []*Service {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*Service, 0, len(s.services))
	for _, svc := range s.services {
		result = append(result, svc)
	}
	return result
}

// Endpoint CRUD
func (s *Store) GetEndpoint(name string) (*Endpoint, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ep, ok := s.endpoints[name]
	return ep, ok
}

func (s *Store) PutEndpoint(ep *Endpoint) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.endpoints[ep.Metadata.Name] = ep
}

func (s *Store) DeleteEndpoint(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.endpoints, name)
}

func (s *Store) ListEndpoints() []*Endpoint {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*Endpoint, 0, len(s.endpoints))
	for _, ep := range s.endpoints {
		result = append(result, ep)
	}
	return result
}

// TODO: Implement FindServicesByOwner, FindEndpointsByOwner
