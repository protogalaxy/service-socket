// Code generated by protoc-gen-go.
// source: messagebroker.proto
// DO NOT EDIT!

/*
Package messagebroker is a generated protocol buffer package.

It is generated from these files:
	messagebroker.proto

It has these top-level messages:
	RouteRequest
	RouteReply
*/
package messagebroker

import proto "github.com/protogalaxy/service-socket/Godeps/_workspace/src/github.com/golang/protobuf/proto"

import (
	context "github.com/protogalaxy/service-socket/Godeps/_workspace/src/golang.org/x/net/context"
	grpc "github.com/protogalaxy/service-socket/Godeps/_workspace/src/google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal

type RouteRequest struct {
	Data []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (m *RouteRequest) Reset()         { *m = RouteRequest{} }
func (m *RouteRequest) String() string { return proto.CompactTextString(m) }
func (*RouteRequest) ProtoMessage()    {}

type RouteReply struct {
}

func (m *RouteReply) Reset()         { *m = RouteReply{} }
func (m *RouteReply) String() string { return proto.CompactTextString(m) }
func (*RouteReply) ProtoMessage()    {}

func init() {
}

// Client API for Broker service

type BrokerClient interface {
	Route(ctx context.Context, in *RouteRequest, opts ...grpc.CallOption) (*RouteReply, error)
}

type brokerClient struct {
	cc *grpc.ClientConn
}

func NewBrokerClient(cc *grpc.ClientConn) BrokerClient {
	return &brokerClient{cc}
}

func (c *brokerClient) Route(ctx context.Context, in *RouteRequest, opts ...grpc.CallOption) (*RouteReply, error) {
	out := new(RouteReply)
	err := grpc.Invoke(ctx, "/messagebroker.Broker/Route", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Broker service

type BrokerServer interface {
	Route(context.Context, *RouteRequest) (*RouteReply, error)
}

func RegisterBrokerServer(s *grpc.Server, srv BrokerServer) {
	s.RegisterService(&_Broker_serviceDesc, srv)
}

func _Broker_Route_Handler(srv interface{}, ctx context.Context, buf []byte) (proto.Message, error) {
	in := new(RouteRequest)
	if err := proto.Unmarshal(buf, in); err != nil {
		return nil, err
	}
	out, err := srv.(BrokerServer).Route(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

var _Broker_serviceDesc = grpc.ServiceDesc{
	ServiceName: "messagebroker.Broker",
	HandlerType: (*BrokerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Route",
			Handler:    _Broker_Route_Handler,
		},
	},
	Streams: []grpc.StreamDesc{},
}
