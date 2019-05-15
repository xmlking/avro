package rpc

import (
	"io"

	"github.com/hamba/avro"
	"golang.org/x/xerrors"
)

// Response represents an Avro response from a Request.
type Response struct {
	// Error determines if this is an error response.
	Error bool

	// Body represents the responses's payload in binary format.
	Body io.Reader

	meta Metadata
}

// Metadata returns the request's metadata.
func (r *Response) Metadata() Metadata {
	return r.meta
}

// Write writes an Avro response in wire format.
func (r *Response) Write(w io.Writer) error {
	wr, ok := w.(*avro.Writer)
	if !ok {
		wr = avro.NewWriter(w, defaultBufSize)
	}

	wr.WriteVal(metadataSchema, r.meta)
	wr.WriteBool(r.Error)
	err := wr.Flush()
	if err != nil {
		return err
	}

	_, err = io.Copy(w, r.Body)
	return err
}

// ReadResponse reads and parses an incoming response from b.
func ReadResponse(b io.Reader) (*Response, error) {
	r, ok := b.(*avro.Reader)
	if !ok {
		r = avro.NewReader(b, defaultBufSize)
	}

	resp := &Response{}

	r.ReadVal(metadataSchema, resp.meta)
	if r.Error != nil {
		return nil, xerrors.Errorf("rpc: error reading response metadata: %w", r.Error)
	}

	resp.Error = r.ReadBool()
	if r.Error != nil {
		return nil, xerrors.Errorf("rpc: error reading response error flag: %w", r.Error)
	}

	resp.Body = r

	return resp, nil
}
