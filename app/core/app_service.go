package core

type AppService struct {
	store *Store

	nodes map[string]*Node
}
