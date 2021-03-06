/*
 *
 * Copyright 2014, Google Inc.
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are
 * met:
 *
 *     * Redistributions of source code must retain the above copyright
 * notice, this list of conditions and the following disclaimer.
 *     * Redistributions in binary form must reproduce the above
 * copyright notice, this list of conditions and the following disclaimer
 * in the documentation and/or other materials provided with the
 * distribution.
 *     * Neither the name of Google Inc. nor the names of its
 * contributors may be used to endorse or promote products derived from
 * this software without specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
 * "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
 * LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
 * A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
 * OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
 * SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
 * LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
 * DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
 * THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 * (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
 * OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 *
 */

package grpc

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
	"strings"
	"sync"

	"github.com/protogalaxy/service-socket/Godeps/_workspace/src/github.com/golang/protobuf/proto"
	"github.com/protogalaxy/service-socket/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/protogalaxy/service-socket/Godeps/_workspace/src/google.golang.org/grpc/codes"
	"github.com/protogalaxy/service-socket/Godeps/_workspace/src/google.golang.org/grpc/metadata"
	"github.com/protogalaxy/service-socket/Godeps/_workspace/src/google.golang.org/grpc/transport"
)

type methodHandler func(srv interface{}, ctx context.Context, buf []byte) (proto.Message, error)

// MethodDesc represents an RPC service's method specification.
type MethodDesc struct {
	MethodName string
	Handler    methodHandler
}

// ServiceDesc represents an RPC service's specification.
type ServiceDesc struct {
	ServiceName string
	// The pointer to the service interface. Used to check whether the user
	// provided implementation satisfies the interface requirements.
	HandlerType interface{}
	Methods     []MethodDesc
	Streams     []StreamDesc
}

// service consists of the information of the server serving this service and
// the methods in this service.
type service struct {
	server interface{} // the server for service methods
	md     map[string]*MethodDesc
	sd     map[string]*StreamDesc
}

// Server is a gRPC server to serve RPC requests.
type Server struct {
	opts  options
	mu    sync.Mutex
	lis   map[net.Listener]bool
	conns map[transport.ServerTransport]bool
	m     map[string]*service // service name -> service info
}

type options struct {
	maxConcurrentStreams uint32
}

// A ServerOption sets options.
type ServerOption func(*options)

// MaxConcurrentStreams returns an Option that will apply a limit on the number
// of concurrent streams to each ServerTransport.
func MaxConcurrentStreams(n uint32) ServerOption {
	return func(o *options) {
		o.maxConcurrentStreams = n
	}
}

// NewServer creates a gRPC server which has no service registered and has not
// started to accept requests yet.
func NewServer(opt ...ServerOption) *Server {
	var opts options
	for _, o := range opt {
		o(&opts)
	}
	return &Server{
		lis:   make(map[net.Listener]bool),
		opts:  opts,
		conns: make(map[transport.ServerTransport]bool),
		m:     make(map[string]*service),
	}
}

// RegisterService register a service and its implementation to the gRPC
// server. Called from the IDL generated code. This must be called before
// invoking Serve.
func (s *Server) RegisterService(sd *ServiceDesc, ss interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Does some sanity checks.
	if _, ok := s.m[sd.ServiceName]; ok {
		log.Fatalf("grpc: Server.RegisterService found duplicate service registration for %q", sd.ServiceName)
	}
	ht := reflect.TypeOf(sd.HandlerType).Elem()
	st := reflect.TypeOf(ss)
	if !st.Implements(ht) {
		log.Fatalf("grpc: Server.RegisterService found the handler of type %v that does not satisfy %v", st, ht)
	}
	srv := &service{
		server: ss,
		md:     make(map[string]*MethodDesc),
		sd:     make(map[string]*StreamDesc),
	}
	for i := range sd.Methods {
		d := &sd.Methods[i]
		srv.md[d.MethodName] = d
	}
	for i := range sd.Streams {
		d := &sd.Streams[i]
		srv.sd[d.StreamName] = d
	}
	s.m[sd.ServiceName] = srv
}

var (
	// ErrServerStopped indicates that the operation is now illegal because of
	// the server being stopped.
	ErrServerStopped = errors.New("grpc: the server has been stopped")
)

// Serve accepts incoming connections on the listener lis, creating a new
// ServerTransport and service goroutine for each. The service goroutines
// read gRPC request and then call the registered handlers to reply to them.
// Service returns when lis.Accept fails.
func (s *Server) Serve(lis net.Listener) error {
	s.mu.Lock()
	if s.lis == nil {
		s.mu.Unlock()
		return ErrServerStopped
	}
	s.lis[lis] = true
	s.mu.Unlock()
	defer func() {
		lis.Close()
		s.mu.Lock()
		delete(s.lis, lis)
		s.mu.Unlock()
	}()
	for {
		c, err := lis.Accept()
		if err != nil {
			return err
		}

		s.mu.Lock()
		if s.conns == nil {
			s.mu.Unlock()
			c.Close()
			return nil
		}
		st, err := transport.NewServerTransport("http2", c, s.opts.maxConcurrentStreams)
		if err != nil {
			s.mu.Unlock()
			c.Close()
			log.Println("grpc: Server.Serve failed to create ServerTransport: ", err)
			continue
		}
		s.conns[st] = true
		s.mu.Unlock()

		go func() {
			st.HandleStreams(func(stream *transport.Stream) {
				s.handleStream(st, stream)
			})
			s.mu.Lock()
			delete(s.conns, st)
			s.mu.Unlock()
		}()
	}
}

func (s *Server) sendProto(t transport.ServerTransport, stream *transport.Stream, msg proto.Message, pf payloadFormat, opts *transport.Options) error {
	p, err := encode(msg, pf)
	if err != nil {
		// This typically indicates a fatal issue (e.g., memory
		// corruption or hardware faults) the application program
		// cannot handle.
		//
		// TODO(zhaoq): There exist other options also such as only closing the
		// faulty stream locally and remotely (Other streams can keep going). Find
		// the optimal option.
		log.Fatalf("grpc: Server failed to encode proto message %v", err)
	}
	return t.Write(stream, p, opts)
}

func (s *Server) processUnaryRPC(t transport.ServerTransport, stream *transport.Stream, srv *service, md *MethodDesc) {
	p := &parser{s: stream}
	for {
		pf, req, err := p.recvMsg()
		if err == io.EOF {
			// The entire stream is done (for unary RPC only).
			return
		}
		if err != nil {
			switch err := err.(type) {
			case transport.ConnectionError:
				// Nothing to do here.
			case transport.StreamError:
				if err := t.WriteStatus(stream, err.Code, err.Desc); err != nil {
					log.Printf("grpc: Server.processUnaryRPC failed to write status: %v", err)
				}
			default:
				panic(fmt.Sprintf("grpc: Unexpected error (%T) from recvMsg: %v", err, err))
			}
			return
		}
		switch pf {
		case compressionNone:
			reply, appErr := md.Handler(srv.server, stream.Context(), req)
			if appErr != nil {
				if err := t.WriteStatus(stream, convertCode(appErr), appErr.Error()); err != nil {
					log.Printf("grpc: Server.processUnaryRPC failed to write status: %v", err)
				}
				return
			}
			opts := &transport.Options{
				Last:  true,
				Delay: false,
			}
			statusCode := codes.OK
			statusDesc := ""
			if err := s.sendProto(t, stream, reply, compressionNone, opts); err != nil {
				if _, ok := err.(transport.ConnectionError); ok {
					return
				}
				if e, ok := err.(transport.StreamError); ok {
					statusCode = e.Code
					statusDesc = e.Desc
				} else {
					statusCode = codes.Unknown
					statusDesc = err.Error()
				}
			}
			if err := t.WriteStatus(stream, statusCode, statusDesc); err != nil {
				log.Printf("grpc: Server.processUnaryRPC failed to write status: %v", err)
			}
		default:
			panic(fmt.Sprintf("payload format to be supported: %d", pf))
		}
	}
}

func (s *Server) processStreamingRPC(t transport.ServerTransport, stream *transport.Stream, srv *service, sd *StreamDesc) {
	ss := &serverStream{
		t: t,
		s: stream,
		p: &parser{s: stream},
	}
	if err := sd.Handler(srv.server, ss); err != nil {
		ss.statusCode = convertCode(err)
		ss.statusDesc = err.Error()
	}
	if err := t.WriteStatus(ss.s, ss.statusCode, ss.statusDesc); err != nil {
		log.Printf("grpc: Server.processStreamingRPC failed to write status: %v", err)
	}
}

func (s *Server) handleStream(t transport.ServerTransport, stream *transport.Stream) {
	sm := stream.Method()
	if sm != "" && sm[0] == '/' {
		sm = sm[1:]
	}
	pos := strings.LastIndex(sm, "/")
	if pos == -1 {
		if err := t.WriteStatus(stream, codes.InvalidArgument, fmt.Sprintf("malformed method name: %q", stream.Method())); err != nil {
			log.Printf("grpc: Server.handleStream failed to write status: %v", err)
		}
		return
	}
	service := sm[:pos]
	method := sm[pos+1:]
	srv, ok := s.m[service]
	if !ok {
		if err := t.WriteStatus(stream, codes.Unimplemented, fmt.Sprintf("unknown service %v", service)); err != nil {
			log.Printf("grpc: Server.handleStream failed to write status: %v", err)
		}
		return
	}
	// Unary RPC or Streaming RPC?
	if md, ok := srv.md[method]; ok {
		s.processUnaryRPC(t, stream, srv, md)
		return
	}
	if sd, ok := srv.sd[method]; ok {
		s.processStreamingRPC(t, stream, srv, sd)
		return
	}
	if err := t.WriteStatus(stream, codes.Unimplemented, fmt.Sprintf("unknown method %v", method)); err != nil {
		log.Printf("grpc: Server.handleStream failed to write status: %v", err)
	}
}

// Stop stops the gRPC server. Once Stop returns, the server stops accepting
// connection requests and closes all the connected connections.
func (s *Server) Stop() {
	s.mu.Lock()
	listeners := s.lis
	s.lis = nil
	cs := s.conns
	s.conns = nil
	s.mu.Unlock()
	for lis := range listeners {
		lis.Close()
	}
	for c := range cs {
		c.Close()
	}
}

// TestingCloseConns closes all exiting transports but keeps s.lis accepting new
// connections. This is for test only now.
func (s *Server) TestingCloseConns() {
	s.mu.Lock()
	for c := range s.conns {
		c.Close()
		delete(s.conns, c)
	}
	s.mu.Unlock()
}

// SendHeader sends header metadata. It may be called at most once from a unary
// RPC handler. The ctx is the RPC handler's Context or one derived from it.
func SendHeader(ctx context.Context, md metadata.MD) error {
	if md.Len() == 0 {
		return nil
	}
	stream, ok := transport.StreamFromContext(ctx)
	if !ok {
		return fmt.Errorf("grpc: failed to fetch the stream from the context %v", ctx)
	}
	t := stream.ServerTransport()
	if t == nil {
		log.Fatalf("grpc: SendHeader: %v has no ServerTransport to send header metadata.", stream)
	}
	return t.WriteHeader(stream, md)
}

// SetTrailer sets the trailer metadata that will be sent when an RPC returns.
// It may be called at most once from a unary RPC handler. The ctx is the RPC
// handler's Context or one derived from it.
func SetTrailer(ctx context.Context, md metadata.MD) error {
	if md.Len() == 0 {
		return nil
	}
	stream, ok := transport.StreamFromContext(ctx)
	if !ok {
		return fmt.Errorf("grpc: failed to fetch the stream from the context %v", ctx)
	}
	return stream.SetTrailer(md)
}
