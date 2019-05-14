package rpc

import (
	"io"
	"io/ioutil"

	"github.com/hamba/avro"
	"golang.org/x/xerrors"
)

// Response represents an Avro response from a Request.
type Response struct {
	// Error determines if this is an error response.
	Error bool

	// Body represents the responses's payload in binary format.
	Body io.ReadCloser

	// Metadata is the response's metadata.
	Metadata map[string][]byte
}

// ReadResponse reads and parses an incoming response from b.
func ReadResponse(b io.Reader) (*Response, error) {
	r := avro.NewReader(b, defaultBufSize)

	resp := &Response{}

	r.ReadVal(metadataSchema, resp.Metadata)
	if r.Error != nil {
		return nil, xerrors.Errorf("rpc: error reading response metadata: %w", r.Error)
	}

	resp.Error = r.ReadBool()
	if r.Error != nil {
		return nil, xerrors.Errorf("rpc: error reading response error flag: %w", r.Error)
	}

	resp.Body = ioutil.NopCloser(r)

	return resp, nil
}
