package core

import (
	"context"
	"fmt"
	"sync"
)

// DatabaseFactory creates a new database provider instance
type DatabaseFactory func(config *DatabaseConfig) (DatabaseProvider, error)

// StorageFactory creates a new storage provider instance
type StorageFactory func(ctx context.Context, config *StorageConfig) (StorageProvider, error)

// Registry holds all registered database and storage providers
type Registry struct {
	databases map[string]DatabaseFactory
	storages  map[string]StorageFactory
	mu        sync.RWMutex
}

var (
	// DefaultRegistry is the global registry instance
	DefaultRegistry = NewRegistry()
)

// NewRegistry creates a new registry
func NewRegistry() *Registry {
	return &Registry{
		databases: make(map[string]DatabaseFactory),
		storages:  make(map[string]StorageFactory),
	}
}

// RegisterDatabase registers a database provider factory
func (r *Registry) RegisterDatabase(name string, factory DatabaseFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.databases[name] = factory
}

// RegisterStorage registers a storage provider factory
func (r *Registry) RegisterStorage(name string, factory StorageFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.storages[name] = factory
}

// GetDatabase creates a database provider instance
func (r *Registry) GetDatabase(name string, config *DatabaseConfig) (DatabaseProvider, error) {
	r.mu.RLock()
	factory, exists := r.databases[name]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("database provider '%s' not registered", name)
	}

	return factory(config)
}

// GetStorage creates a storage provider instance
func (r *Registry) GetStorage(ctx context.Context, name string, config *StorageConfig) (StorageProvider, error) {
	r.mu.RLock()
	factory, exists := r.storages[name]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("storage provider '%s' not registered", name)
	}

	return factory(ctx, config)
}

// ListDatabases returns a list of registered database provider names
func (r *Registry) ListDatabases() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.databases))
	for name := range r.databases {
		names = append(names, name)
	}
	return names
}

// ListStorages returns a list of registered storage provider names
func (r *Registry) ListStorages() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.storages))
	for name := range r.storages {
		names = append(names, name)
	}
	return names
}

// RegisterDatabase registers a database provider in the default registry
func RegisterDatabase(name string, factory DatabaseFactory) {
	DefaultRegistry.RegisterDatabase(name, factory)
}

// RegisterStorage registers a storage provider in the default registry
func RegisterStorage(name string, factory StorageFactory) {
	DefaultRegistry.RegisterStorage(name, factory)
}

// GetDatabase creates a database provider from the default registry
func GetDatabase(name string, config *DatabaseConfig) (DatabaseProvider, error) {
	return DefaultRegistry.GetDatabase(name, config)
}

// GetStorage creates a storage provider from the default registry
func GetStorage(ctx context.Context, name string, config *StorageConfig) (StorageProvider, error) {
	return DefaultRegistry.GetStorage(ctx, name, config)
}

// ListDatabases returns registered database providers from the default registry
func ListDatabases() []string {
	return DefaultRegistry.ListDatabases()
}

// ListStorages returns registered storage providers from the default registry
func ListStorages() []string {
	return DefaultRegistry.ListStorages()
}
