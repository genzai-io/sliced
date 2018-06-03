package api

// Cluster wide registry of schemas and raft services.
type Store interface {
	Raft() RaftService

	//
	CreateDatabase()

	DeleteDatabase()

	Database() []Database

	GetSchema(id int32) Database
}
