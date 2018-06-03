package api

type ConnKind byte

const (
	ConnCommand ConnKind = 0
	ConnPubSub  ConnKind = 1
	ConnRaft    ConnKind = 2
	ConnQueue   ConnKind = 3
)

type Conn interface {
	Kind() ConnKind

	SetKind(kind ConnKind)

	//
	Close() error

	//
	Handler() IHandler

	// Redirect commands to a different Handler
	SetHandler(handler IHandler) IHandler

	//
	Durability() Durability

	//
	Raft() RaftService

	//
	SetRaft(raft RaftService)
}
