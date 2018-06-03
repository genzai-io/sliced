package core

import (
	"github.com/genzai-io/sliced"
	"github.com/genzai-io/sliced/app/api"
	"github.com/genzai-io/sliced/app/fs"
	"github.com/genzai-io/sliced/common/service"
)

var Instance *Service

func init() {
	Instance = NewService()
}

// Instance store for the definitions of system objects and Cluster topology.
// This maintains the Cluster wide state including re-balancing, adding slices,
// removing slices, adding nodes, removing nodes, etc. It also keeps track of
// Cluster wide object definitions that will implicitly be available on every
// slice without having to define it on every slice. This isn't to be confused
// with slice level definitions which allow for more granular object definitions.
type Service struct {
	service.BaseService

	Schema  *Store
	Cluster *ClusterService
	Drives  *fs.DriveService
}

func NewService() *Service {
	s := &Service{}

	s.BaseService = *service.NewBaseService(moved.Logger, "core", s)

	return s
}

func (b *Service) OnStart() error {
	var err error
	// Start schema service
	b.Schema = newSchema()
	err = b.Schema.Start()
	if err != nil {
		return err
	}
	//api.Database = b.Database

	// Start drive service
	b.Drives = fs.NewDriveService()
	err = b.Drives.Start()
	if err != nil {
		b.Schema.Stop()
		return err
	}
	api.Drives = b.Drives

	// Start Cluster
	b.Cluster = newCluster(b.Schema)
	err = b.Cluster.Start()
	if err != nil {
		b.Schema.Stop()
		b.Drives.Stop()
		return err
	}
	api.Cluster = b.Cluster

	return nil
}

func (b *Service) OnStop() {
	if err := b.Cluster.Stop(); err != nil {
		b.Logger.Error().AnErr("err", err).Msg("Cluster.Stop() error")
	}

	if err := b.Drives.Stop(); err != nil {
		b.Logger.Error().AnErr("err", err).Msg("Drives.Stop() error")
	}

	if err := b.Schema.Stop(); err != nil {
		b.Logger.Error().AnErr("err", err).Msg("Database.Stop() error")
	}
}
