package rpc

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/hamba/avro"
)

type Request struct {
	Message string

	Body io.ReadCloser

	Metadata map[string]string

	RemoteAddr string

	ctx context.Context
}

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
		Message: name,
		Body:    body,
	}, nil
}

func (r *Request) Context() context.Context {
	if r.ctx != nil {
		return r.ctx
	}
	return context.Background()
}

func (r *Request) WithContext(ctx context.Context) *Request {
	if ctx == nil {
		panic("nil context")
	}

	r2 := new(Request)
	*r2 = *r
	r2.ctx = ctx

	return r2
}

func ReadRequest(b bufio.Reader) (*Request, error) {
	// Read Metadata

	// Read message name

	// Set body reader

	return nil, nil
}
