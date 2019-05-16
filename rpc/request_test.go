package rpc_test

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/hamba/avro/rpc"
	"github.com/stretchr/testify/assert"
)

func TestNewRequest(t *testing.T) {
	req, err := rpc.NewRequest(echoProtocol, "echo", echoRequest{Text: "foo"})

	assert.NoError(t, err)
	assert.Equal(t, "echo", req.Name)
	body, _ := ioutil.ReadAll(req.Body)
	assert.Equal(t, []byte{0x06, 0x66, 0x6f, 0x6f}, body)
}

func TestNewRequest_BadMessage(t *testing.T) {
	_, err := rpc.NewRequest(echoProtocol, "test", echoRequest{Text: "foo"})

	assert.Error(t, err)
}

func TestNewRequest_BadValue(t *testing.T) {
	_, err := rpc.NewRequest(echoProtocol, "echo", "foo")

	assert.Error(t, err)
}

func TestRequest_Write(t *testing.T) {
	req, _ := rpc.NewRequest(echoProtocol, "echo", echoRequest{Text: "foo"})
	req.Metadata().Set("foo", []byte("foo"))

	buf := bytes.NewBuffer(nil)
	err := req.Write(buf)

	want := []byte{0x01, 0x10, 0x06, 0x66, 0x6F, 0x6F, 0x06, 0x66, 0x6F, 0x6F, 0x00, 0x08, 0x65, 0x63, 0x68, 0x6f, 0x06, 0x66, 0x6f, 0x6f}
	assert.NoError(t, err)
	assert.Equal(t, want, buf.Bytes())
}

func TestReadRequest(t *testing.T) {
	data := []byte{0x01, 0x10, 0x06, 0x66, 0x6F, 0x6F, 0x06, 0x66, 0x6F, 0x6F, 0x00, 0x08, 0x65, 0x63, 0x68, 0x6f, 0x06, 0x66, 0x6f, 0x6f}

	req, err := rpc.ReadRequest(bytes.NewReader(data))

	assert.NoError(t, err)
	assert.Equal(t,rpc.Metadata{"foo": []byte("foo")}, req.Metadata())
	assert.Equal(t, "echo", req.Name)
	body, _ := ioutil.ReadAll(req.Body)
	assert.Equal(t, []byte{0x06, 0x66, 0x6f, 0x6f}, body)
}
