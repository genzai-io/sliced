package server

import (
	"context"

	api_pb "github.com/genzai-io/sliced/proto/api"
)

//
//
//
func (s *Web) Auth(ctx context.Context, req *api_pb.AuthRequest) (*api_pb.AuthReply, error) {
	return nil, nil
}

//
//
//
func (s *Web) Register(ctx context.Context, req *api_pb.RegisterRequest) (*api_pb.RegisterReply, error) {
	return nil, nil
}

//
//
//
func (s *Web) Events(req *api_pb.EventsRequest, srv api_pb.APIService_EventsServer) error {
	return nil
}
