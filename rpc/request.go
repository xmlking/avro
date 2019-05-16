package rpc

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/hamba/avro"
	"golang.org/x/xerrors"
)

// Request represents the Avro request received by a server
// or to be sent by a client.
type Request struct {
	// Protocol is the protocol being used to sent the message.
	Protocol *avro.Protocol

	// Message is the message being sent.
	Message *avro.Message

	// Name is the message name.
	Name string

	// Body represents the request's payload in binary format.
	Body io.Reader

	// RemoteAddr is the network address that sent the request.
	// This field is ignored by the client.
	RemoteAddr string

	meta Metadata
	ctx  context.Context
}

// NewRequest returns a new Request with the given message name and encoded body data.
func NewRequest(protocol *avro.Protocol, name string, v interface{}) (*Request, error) {
	msg := protocol.Message(name)
	if msg == nil {
		return nil, fmt.Errorf("rpc: no message with name %s", name)
	}

	b, err := avro.Marshal(msg.Request(), v)
	if err != nil {
		return nil, err
	}
	body := bytes.NewReader(b)

	return &Request{
		Protocol: protocol,
		Message:  msg,
		Name:     name,
		Body:     body,
		meta:     Metadata{},
	}, nil
}

// Context returns the request's context.
func (r *Request) Context() context.Context {
	if r.ctx != nil {
		return r.ctx
	}

	return context.Background()
}

// WithContext returns a shallow copy of r with its context changed
// to ctx. The provided ctx must be non-nil.
func (r *Request) WithContext(ctx context.Context) *Request {
	if ctx == nil {
		panic("rpc: nil context")
	}

	r2 := new(Request)
	*r2 = *r
	r2.ctx = ctx

	return r2
}

// Metadata returns the request's metadata.
func (r *Request) Metadata() Metadata {
	return r.meta
}

// Write writes an Avro request in wire format.
func (r *Request) Write(w io.Writer) error {
	wr, ok := w.(*avro.Writer)
	if !ok {
		wr = avro.NewWriter(w, defaultBufSize)
	}

	wr.WriteVal(metadataSchema, r.meta)
	wr.WriteString(r.Name)
	err := wr.Flush()
	if err != nil {
		return err
	}

	_, err = io.Copy(w, r.Body)
	return err
}

// ReadRequest reads and parses an incoming request from b.
func ReadRequest(r io.Reader) (*Request, error) {
	rdr, ok := r.(*avro.Reader)
	if !ok {
		rdr = avro.NewReader(r, defaultBufSize)
	}

	req := &Request{}

	rdr.ReadVal(metadataSchema, &req.meta)
	if rdr.Error != nil {
		return nil, xerrors.Errorf("rpc: error reading request metadata: %w", rdr.Error)
	}

	req.Name = rdr.ReadString()
	if rdr.Error != nil {
		return nil, xerrors.Errorf("rpc: error reading request message name: %w", rdr.Error)
	}

	req.Body = rdr

	return req, nil
}
