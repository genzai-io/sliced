package api

// Cluster wide registry of schemas and Raft services.
type Store interface {
	Raft() RaftService

	//
	CreateDatabase()

	DeleteDatabase()

	Database() []Database

	GetSchema(id int32) Database
}
