package rpc

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/hamba/avro"
	"golang.org/x/xerrors"
)

// Request represents the Avro request received by a server
// or to be sent by a client.
type Request struct {
	// Message specifies the message name.
	Message string

	// Body represents the request's payload in binary format.
	Body io.ReadCloser

	// Metadata is the request's metadata.
	Metadata map[string][]byte

	// RemoteAddr is the network address that sent the request.
	// This field is ignored by the client.
	RemoteAddr string

	ctx context.Context
}

// NewRequest returns a new Request with the given message name and encoded body data.
func NewRequest(proto *avro.Protocol, name string, v interface{}) (*Request, error) {
	msg := proto.Message(name)
	if msg == nil {
		return nil, fmt.Errorf("rpc: no message with name %s", name)
	}

	b, err := avro.Marshal(msg.Request(), v)
	if err != nil {
		return nil, err
	}
	body := ioutil.NopCloser(bytes.NewReader(b))

	return &Request{
		Message:  name,
		Body:     body,
		Metadata: map[string][]byte{},
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

// Write writes an Avro request in wire format.
func (r *Request) Write(w io.Writer) error {
	wr := avro.NewWriter(w, defaultBufSize)

	wr.WriteVal(metadataSchema, r.Metadata)
	wr.WriteString(r.Message)
	err := wr.Flush()
	if err != nil {
		return err
	}

	_, err = io.Copy(w, r.Body)
	if err != nil {
		return err
	}

	return r.Body.Close()
}

// ReadRequest reads and parses an incoming request from b.
func ReadRequest(b io.Reader) (*Request, error) {
	r := avro.NewReader(b, defaultBufSize)

	req := &Request{}

	r.ReadVal(metadataSchema, &req.Metadata)
	if r.Error != nil {
		return nil, xerrors.Errorf("rpc: error reading request metadata: %w", r.Error)
	}

	req.Message = r.ReadString()
	if r.Error != nil {
		return nil, xerrors.Errorf("rpc: error reading request message name: %w", r.Error)
	}

	req.Body = ioutil.NopCloser(r)

	return req, nil
}
