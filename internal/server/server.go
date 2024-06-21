package server

import (
	"context"

	api "github.com/innazh/proglog/api/v1"

	"google.golang.org/grpc"
)

/*
By defining the CommitLog interface, we decouple a concerete Log implementation from our service.
That way our service can use any structure that satisfies the CommitLog interface.
Now we can use an in-memory Log and use it for testing, for example. While still using the disk-persistent storage for prod.
*/
type CommitLog interface { //internal/log & internal/segment will already have this interface
	Append(*api.Record) (uint64, error)
	Read(uint64) (*api.Record, error)
}

type Config struct {
	CommitLog CommitLog
}

/*
In order to implement our Log Service, we need to define a struct which implements LogServer's interface. This struct is grpcServer.

The instructions also state that for forward compativility, we need to implement the api.UnimplementedLogServer.
*/

var _ api.LogServer = (*grpcServer)(nil)

type grpcServer struct {
	api.UnimplementedLogServer
	*Config
}

func newgrpcServer(config *Config) (srv *grpcServer, err error) {
	srv = &grpcServer{Config: config}
	return srv, nil
}
func NewGRPCServer(config *Config) (*grpc.Server, error) {
	gsrv := grpc.NewServer()
	srv, err := newgrpcServer(config)
	if err != nil {
		return nil, err
	}
	api.RegisterLogServer(gsrv, srv)
	return gsrv, nil
}

func (s *grpcServer) Produce(ctx context.Context, req *api.ProduceRequest) (*api.ProduceResponse, error) {
	offset, err := s.CommitLog.Append(req.Record)
	if err != nil {
		return nil, err
	}
	return &api.ProduceResponse{Offset: offset}, nil
}

func (s *grpcServer) Consume(ctx context.Context, req *api.ConsumeRequest) (*api.ConsumeResponse, error) {
	record, err := s.CommitLog.Read(req.Offset)
	if err != nil {
		return nil, err
	}
	return &api.ConsumeResponse{Record: record}, nil
}

/*
ProduceStream implements bidirectional streaming RPC.
It's bidirectional so that the client can get a response back on each one of its requests to add log to the server.
*/
func (s *grpcServer) ProduceStream(stream api.Log_ProduceStreamServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			return err
		}
		res, err := s.Produce(stream.Context(), req)
		if err != nil {
			return err
		}
		if err = stream.Send(res); err != nil {
			return err
		}
	}
}

/*
ConsumeStream implements a server-side streaming RPC.
This allows the client to request where the server needs to start reading and the server will stream every record that follows it.
The server will stream all newly added records, until the stream is closed.
*/
func (s *grpcServer) ConsumeStream(req *api.ConsumeRequest, stream api.Log_ConsumeStreamServer) error {
	for {
		select {
		case <-stream.Context().Done(): //Allows the server to stop streaming if the client cancels the request or if a context timeout occurs.
			return nil
		default:
			res, err := s.Consume(stream.Context(), req)
			switch err.(type) {
			case nil:
			case api.ErrOffsetOutOfRange:
				continue
			default:
				return err
			}
			if err = stream.Send(res); err != nil {
				return err
			}
			req.Offset++
		}
	}
}
