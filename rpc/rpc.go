package rpc

import "github.com/hamba/avro"

const (
	defaultBufSize = 1024
)

var (
	metadataSchema = avro.MustParse(`{"type":"map", "values": "bytes"}`)

	HandshakeRequestSchema = avro.MustParse(`{
  "type": "record",
  "name": "HandshakeRequest", "namespace":"org.apache.avro.ipc",
  "fields": [
    {"name": "clientHash", "type": {"type": "fixed", "name": "MD5", "size": 16}},
    {"name": "clientProtocol", "type": ["null", "string"]},
    {"name": "serverHash", "type": "MD5"},
    {"name": "meta", "type": ["null", {"type": "map", "values": "bytes"}]}
  ]
}`)
	HandshakeResponseSchema = avro.MustParse(`{
  "type": "record",
  "name": "HandshakeResponse", "namespace": "org.apache.avro.ipc",
  "fields": [
    {"name": "match", "type": {"type": "enum", "name": "HandshakeMatch", "symbols": ["BOTH", "CLIENT", "NONE"]}},
    {"name": "serverProtocol", "type": ["null", "string"]},
    {"name": "serverHash", "type": ["null", {"type": "fixed", "name": "MD5", "size": 16}]},
    {"name": "meta", "type": ["null", {"type": "map", "values": "bytes"}]}
  ]
}`)
)

// HandshakeRequest represents an Avro request handshake.
type HandshakeRequest struct {
	ClientHash     [16]byte           `avro:"clientHash"`
	ClientProtocol *string            `avro:"clientProtocol"`
	ServerHash     [16]byte           `avro:"serverHash"`
	Metadata       *map[string][]byte `avro:"meta"`
}

// Match constants.
const (
	MatchBoth   = "BOTH"
	MatchClient = "CLIENT"
	MatchNone   = "NONE"
)

// HandshakeResponse represents an Avro response handshake.
type HandshakeResponse struct {
	Match          string             `avro:"match"`
	ServerProtocol *string            `avro:"serverProtocol"`
	ServerHash     [16]byte           `avro:"serverHash"`
	Metadata       *map[string][]byte `avro:"meta"`
}
