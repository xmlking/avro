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

// ReadResponse reads and parses an incoming response from b.
func ReadResponse(r io.Reader) (*Response, error) {
	rdr, ok := r.(*avro.Reader)
	if !ok {
		rdr = avro.NewReader(r, defaultBufSize)
	}

	resp := &Response{}

	rdr.ReadVal(metadataSchema, resp.meta)
	if rdr.Error != nil {
		return nil, xerrors.Errorf("rpc: error reading response metadata: %w", rdr.Error)
	}

	resp.Error = rdr.ReadBool()
	if rdr.Error != nil {
		return nil, xerrors.Errorf("rpc: error reading response error flag: %w", rdr.Error)
	}

	resp.Body = rdr

	return resp, nil
}
