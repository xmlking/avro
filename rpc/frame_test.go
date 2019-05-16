package rpc_test

import (
	"bytes"
	"io"
	"math/rand"
	"testing"

	"github.com/hamba/avro/rpc"
	"github.com/stretchr/testify/assert"
)

func TestNewFrameReader(t *testing.T) {
	r := rpc.NewFrameReader(nil)

	assert.IsType(t, &rpc.FrameReader{}, r)
	assert.Implements(t, (*io.Reader)(nil), r)
}

func TestFrameReader_Read(t *testing.T) {
	br := bytes.NewReader([]byte{0x00, 0x00, 0x00, 0x02, 0x05, 0x06, 0x00, 0x00, 0x00, 0x00})
	r := rpc.NewFrameReader(br)

	buf := make([]byte, 10)
	n, err := r.Read(buf)

	assert.Equal(t, io.EOF, err)
	assert.Equal(t, 2, n)
	assert.Equal(t, []byte{0x05, 0x06}, buf[:n])
}

func TestFrameReader_ReadMoreDataThanSlice(t *testing.T) {
	br := bytes.NewReader([]byte{0x00, 0x00, 0x00, 0x0B, 0x05, 0x06, 0x05, 0x06, 0x05, 0x06, 0x05, 0x06, 0x05, 0x06, 0x05, 0x00, 0x00, 0x00, 0x00})
	r := rpc.NewFrameReader(br)

	buf := make([]byte, 10)
	n, err := r.Read(buf)

	assert.NoError(t, err)
	assert.Equal(t, 10, n)
	assert.Equal(t, []byte{0x05, 0x06, 0x05, 0x06, 0x05, 0x06, 0x05, 0x06, 0x05, 0x06}, buf[:n])
}

func TestFrameReader_ReadHandlesHeaderReadError(t *testing.T) {
	br := bytes.NewReader([]byte{})
	r := rpc.NewFrameReader(br)

	buf := make([]byte, 10)
	_, err := r.Read(buf)

	assert.Error(t, err)
}

func TestFrameReader_ReadHandlesInvalidHeader(t *testing.T) {
	br := bytes.NewReader([]byte{0x05})
	r := rpc.NewFrameReader(br)

	buf := make([]byte, 10)
	_, err := r.Read(buf)

	assert.Error(t, err)
}

func TestFrameReader_ReadHandlesDataReadError(t *testing.T) {
	br := bytes.NewReader([]byte{0x00, 0x00, 0x00, 0x0B})
	r := rpc.NewFrameReader(br)

	buf := make([]byte, 10)
	_, err := r.Read(buf)

	assert.Error(t, err)
}

func TestNewFrameWriter(t *testing.T) {
	r := rpc.NewFrameWriter(nil)

	assert.IsType(t, &rpc.FrameWriter{}, r)
	assert.Implements(t, (*io.Writer)(nil), r)
}

func TestFrameWriter_Write(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	w := rpc.NewFrameWriter(buf)

	n, werr := w.Write([]byte{0x05, 0x06})
	ferr := w.Flush()

	assert.NoError(t, werr)
	assert.NoError(t, ferr)
	assert.Equal(t, 2, n)
	assert.Equal(t, []byte{0x00, 0x00, 0x00, 0x02, 0x05, 0x06, 0x00, 0x00, 0x00, 0x00}, buf.Bytes())
}

func TestFrameWriter_WriteCanWriteLargeSlices(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	w := rpc.NewFrameWriter(buf)
	w.Write([]byte{0x05, 0x06})
	data := make([]byte, 4097)
	rand.Read(data)

	n, werr := w.Write(data)
	ferr := w.Flush()

	got := buf.Bytes()
	assert.NoError(t, werr)
	assert.NoError(t, ferr)
	assert.Equal(t, 4097, n)
	assert.Equal(t, []byte{0x00, 0x00, 0x00, 0x02, 0x05, 0x06}, got[:6])
	assert.Equal(t, []byte{0x0, 0x0, 0x10, 0x01}, got[6:10])
	assert.Equal(t, data, got[10:4107])
	assert.Equal(t, []byte{0x00, 0x00, 0x00, 0x00}, got[4107:])
}

func TestFrameWriter_Flush(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	w := rpc.NewFrameWriter(buf)

	err := w.Flush()

	assert.NoError(t, err)
	assert.Equal(t, []byte{0x00, 0x00, 0x00, 0x00}, buf.Bytes())
}
