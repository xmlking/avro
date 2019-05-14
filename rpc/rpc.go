package rpc

import "github.com/hamba/avro"

const (
	defaultBufSize = 1024
)

var (
	metadataSchema = avro.MustParse(`{"type":"map", "values": "bytes"}`)
)
