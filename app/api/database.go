package api

type Databases interface {
	GetByName(name string) Database

	GetByID(id int32) Database
}

type Database interface {
	Name()

	Slices()

	CreateTopic()

	DeleteTopic()

	CreateQueue()

	DeleteQueue()

	CreateTable()

	DeleteTable()
}
