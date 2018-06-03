package api

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
