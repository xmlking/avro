package rpc

import (
	"context"
	"io"

	"github.com/hamba/avro"
	"github.com/modern-go/concurrent"
)

type Conn interface {
	io.ReadWriter

	RemoteAddr() string

	Protocol() *avro.Protocol
	SetProtocol(*avro.Protocol)
}

type ServerCodec struct {
	protocol *avro.Protocol
	handler  Handler

	protocolCache *concurrent.Map // map[string]*avro.Protocol
}

func NewServerCodec(proto *avro.Protocol, handler Handler) *ServerCodec {
	return &ServerCodec{
		protocol: proto,
		handler:  handler,
	}
}

func (c *ServerCodec) handshake(conn Conn) (*avro.Protocol, error) {
	if p := conn.Protocol(); p != nil {
		return p, nil
	}

	return nil, nil
}

func (c *ServerCodec) Serve(ctx context.Context, conn Conn) error {
	//TODO: Handshake

	req, err := ReadRequest(conn)
	if err != nil {
		return err
	}
	ctx, cancelCtx := context.WithCancel(ctx)
	req.ctx = ctx
	req.RemoteAddr = conn.RemoteAddr()

	//TODO: Make response writer

	if req.Message == "" {
		//TODO: handle pong
	}

	//c.handler.Serve(w, req)
	cancelCtx()

	return nil
}
