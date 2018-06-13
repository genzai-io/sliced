package server

import (
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/go-chi/chi"
	chiMiddleware "github.com/go-chi/chi/middleware"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	_ "github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/genzai-io/sliced"
	"github.com/genzai-io/sliced/assets/gzip"
	"github.com/genzai-io/sliced/assets/ui"
	api_pb "github.com/genzai-io/sliced/proto/api"
	"github.com/genzai-io/sliced/common/service"
	"github.com/rs/cors"
	"google.golang.org/grpc"
)

// Provides HTTP web services
type Web struct {
	service.BaseService
	*grpcweb.WrappedGrpcServer

	mu         sync.Mutex
	host       string
	listener   net.Listener
	addr       *net.TCPAddr
	grpcServer *grpc.Server
	router     chi.Router
	handler    http.Handler
	Err        error
}

func NewWeb(host string) *Web {
	grpcServer := grpc.NewServer()
	wrappedGrpc := grpcweb.WrapServer(grpcServer, grpcweb.WithWebsockets(true))

	server := &Web{
		WrappedGrpcServer: wrappedGrpc,
		host:              host,
		grpcServer:        grpcServer,
		Err:               nil,
	}

	// Register services
	api_pb.RegisterAPIServiceServer(grpcServer, server)

	router := chi.NewRouter()
	router.Use(
		chiMiddleware.RequestLogger(&zeroLogger{}),
		chiMiddleware.Recoverer,
		server.Handler,
		cors.New(cors.Options{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token",},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: true,
			MaxAge:           300, // Maximum value not ignored by any of major browsers
		}).Handler,
	)

	FileServer(router, "/s", ui_data.FS())

	router.NotFound(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(404)
	})

	server.router = router

	server.BaseService = *service.NewBaseService(moved.Logger, "http.server", server)

	return server
}

func (s *Web) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.IsAcceptableGrpcCorsRequest(r) || s.IsGrpcWebRequest(r) {
			s.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Web) Addr() *net.TCPAddr {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.addr
}

func (s *Web) Close() error {
	return s.Stop()
}

func (s *Web) OnStop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.listener != nil {
		s.Err = s.listener.Close()
		s.listener = nil

		s.grpcServer.Stop()
		s.addr = nil
	}
	//s.grpcServer.Stop()
}

func (s *Web) OnStart() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var err error
	s.listener, err = net.Listen("tcp", s.host)
	if err != nil {
		// Let the OS pick a port
		s.listener, err = net.Listen("tcp", ":0")
		if err != nil {
			return err
		}
	}

	s.addr = s.listener.Addr().(*net.TCPAddr)

	go func() {
		defer s.Stop()
		if err := http.Serve(s.listener, s.router); err != nil {
			//grpclog.Fatalf("failed starting http2 server: %v", s.Err)
		}
	}()

	return nil
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	fs := http.StripPrefix(path, gzip.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}

func CheckForUpdate() {
	/*
{
	"url": "https://mycompany.example.com/myapp/releases/myrelease",
	"name": "My Release Name",
	"notes": "Theses are some release notes innit",
	"pub_date": "2013-09-18T12:29:53+01:00"
}
	 */
}
