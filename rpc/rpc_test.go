package rpc_test

import "github.com/hamba/avro"

var echoProtocol = avro.MustParseProtocol(`
{
	"protocol": "test",
	"messages": {
		"echo": {
			"request":[
				{"name":"text","type":"string"}
			],
			"response": "string"
		}
	}
}`)

type echoRequest struct {
	Text string `avro:"text"`
}
