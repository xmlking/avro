package rpc

import (
	"context"
	"io"

	"github.com/hamba/avro"
	"github.com/modern-go/concurrent"
	"golang.org/x/xerrors"
)

type ResponseWriter interface {
	Metadata(key string, value []byte)

	Write(value interface{})

	Error(value interface{})
}

type Handler interface {
	Serve(ResponseWriter, *Request)
}

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

func NewServerCodec(protocol *avro.Protocol, handler Handler) *ServerCodec {
	return &ServerCodec{
		protocol: protocol,
		handler:  handler,
	}
}

func (c *ServerCodec) remoteProtocol(hash [16]byte) *avro.Protocol {
	p, ok := c.protocolCache.Load(hash)
	if ok {
		return p.(*avro.Protocol)
	}

	return nil
}

func (c *ServerCodec) handshake(r *avro.Reader, w *avro.Writer, conn Conn) (*avro.Protocol, error) {
	if p := conn.Protocol(); p != nil {
		return p, nil
	}

	req := &HandshakeRequest{}
	r.ReadVal(HandshakeRequestSchema, req)
	if r.Error != nil {
		return nil, xerrors.Errorf("rpc: error reading handshake request: %w", r.Error)
	}

	remoteProto := c.remoteProtocol(req.ClientHash)
	if remoteProto == nil && req.ClientProtocol != nil {
		proto, err := avro.ParseProtocol(*req.ClientProtocol)
		if err != nil {
			return nil, xerrors.Errorf("rpc: error parsing handshake client protocol: %w", r.Error)
		}

		remoteProto = proto
		c.protocolCache.Store(req.ClientHash, proto)
	}

	resp := HandshakeResponse{
		Match:      MatchNone,
		ServerHash: c.protocol.Hash(),
	}

	if resp.ServerHash == req.ServerHash && remoteProto != nil {
		resp.Match = MatchBoth
	} else if resp.ServerHash != req.ServerHash && remoteProto != nil {
		resp.Match = MatchClient
	}

	if resp.Match != MatchBoth {
		str := c.protocol.String()
		resp.ServerProtocol = &str
	}

	w.WriteVal(HandshakeResponseSchema, resp)

	return remoteProto, nil
}

func (c *ServerCodec) Serve(ctx context.Context, conn Conn) {
	//TODO: pool these?
	r := avro.NewReader(conn, defaultBufSize)
	w := avro.NewWriter(conn, defaultBufSize)

	proto, err := c.handshake(r, w, conn)
	if err != nil || proto == nil {
		if err != nil {
			//TODO: send system error for bad handshake
		}
		return
	}

	req, err := ReadRequest(r)
	if err != nil {
		//TODO: send system error for bad request
		return
	}
	ctx, cancelCtx := context.WithCancel(ctx)
	req.ctx = ctx
	req.Protocol = c.protocol
	req.RemoteAddr = conn.RemoteAddr()

	//TODO: Make response writer

	if req.Name == "" {
		//TODO: handle pong
	}

	//TODO: resolve message

	//c.handler.Serve(w, req)
	cancelCtx()

	w.Flush()
}
