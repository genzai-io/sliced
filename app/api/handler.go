package api

// Default handler
var Handler IHandler

type IHandler interface {
	Parse(ctx *Context) Command

	Commit(ctx *Context)
}

