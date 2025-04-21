package handlers

import (
	"AvitoPVZService/Service/internal/repositories/interfaces"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GrpcHandlers struct {
	Data interfaces.Repository
}

func NewGrpcHandlers(repo interfaces.Repository) *GrpcHandlers {
	return &GrpcHandlers{Data: repo}
}

type GetPVZListRequest struct{}

type GetPVZListResponse struct{ PVZs []*ProtoPVZ }

type ProtoPVZ struct {
	Id               string
	RegistrationDate *timestamppb.Timestamp
	City             string
}

func (g *GrpcHandlers) GetPVZList(ctx context.Context, _ *GetPVZListRequest) (*GetPVZListResponse, error) {
	endPeriod := "2025-04-20 16:30:27.799684"
	startPeriod := "2025-04-19 16:30:27.799684"
	limit := 4
	offset := 0

	err, resp := g.Data.GrpcListPVz(endPeriod, startPeriod, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("GetPVZList query error: %w", err)
	}

	return &GetPVZListResponse{PVZs: resp}, nil
}

var _PVZService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "pvz.v1.PVZService",
	HandlerType: (*GrpcHandlers)(nil),
	Methods:     []grpc.MethodDesc{{MethodName: "GetPVZList", Handler: _PVZService_GetPVZList_Handler}},
	Streams:     []grpc.StreamDesc{},
	Metadata:    "pvz.proto",
}

func RegisterPVZServiceServer(s *grpc.Server, srv *GrpcHandlers) {
	s.RegisterService(&_PVZService_serviceDesc, srv)
}

func _PVZService_GetPVZList_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetPVZListRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(*GrpcHandlers).GetPVZList(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/pvz.v1.PVZService/GetPVZList"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(*GrpcHandlers).GetPVZList(ctx, req.(*GetPVZListRequest))
	}
	return interceptor(ctx, in, info, handler)
}
