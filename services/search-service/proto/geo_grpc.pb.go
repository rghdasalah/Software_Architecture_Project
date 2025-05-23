// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v6.30.2
// source: geo.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	GeoDistanceService_FilterRidesByProximity_FullMethodName = "/proto.GeoDistanceService/FilterRidesByProximity"
)

// GeoDistanceServiceClient is the client API for GeoDistanceService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GeoDistanceServiceClient interface {
	FilterRidesByProximity(ctx context.Context, in *FilterRequest, opts ...grpc.CallOption) (*FilterResponse, error)
}

type geoDistanceServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewGeoDistanceServiceClient(cc grpc.ClientConnInterface) GeoDistanceServiceClient {
	return &geoDistanceServiceClient{cc}
}

func (c *geoDistanceServiceClient) FilterRidesByProximity(ctx context.Context, in *FilterRequest, opts ...grpc.CallOption) (*FilterResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(FilterResponse)
	err := c.cc.Invoke(ctx, GeoDistanceService_FilterRidesByProximity_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GeoDistanceServiceServer is the server API for GeoDistanceService service.
// All implementations must embed UnimplementedGeoDistanceServiceServer
// for forward compatibility.
type GeoDistanceServiceServer interface {
	FilterRidesByProximity(context.Context, *FilterRequest) (*FilterResponse, error)
	mustEmbedUnimplementedGeoDistanceServiceServer()
}

// UnimplementedGeoDistanceServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedGeoDistanceServiceServer struct{}

func (UnimplementedGeoDistanceServiceServer) FilterRidesByProximity(context.Context, *FilterRequest) (*FilterResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FilterRidesByProximity not implemented")
}
func (UnimplementedGeoDistanceServiceServer) mustEmbedUnimplementedGeoDistanceServiceServer() {}
func (UnimplementedGeoDistanceServiceServer) testEmbeddedByValue()                            {}

// UnsafeGeoDistanceServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GeoDistanceServiceServer will
// result in compilation errors.
type UnsafeGeoDistanceServiceServer interface {
	mustEmbedUnimplementedGeoDistanceServiceServer()
}

func RegisterGeoDistanceServiceServer(s grpc.ServiceRegistrar, srv GeoDistanceServiceServer) {
	// If the following call pancis, it indicates UnimplementedGeoDistanceServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&GeoDistanceService_ServiceDesc, srv)
}

func _GeoDistanceService_FilterRidesByProximity_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FilterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GeoDistanceServiceServer).FilterRidesByProximity(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GeoDistanceService_FilterRidesByProximity_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GeoDistanceServiceServer).FilterRidesByProximity(ctx, req.(*FilterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// GeoDistanceService_ServiceDesc is the grpc.ServiceDesc for GeoDistanceService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var GeoDistanceService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.GeoDistanceService",
	HandlerType: (*GeoDistanceServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "FilterRidesByProximity",
			Handler:    _GeoDistanceService_FilterRidesByProximity_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "geo.proto",
}
