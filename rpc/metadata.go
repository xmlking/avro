package rpc

type Metadata map[string][]byte

func (m Metadata) Set(key string, value []byte) {
	m[key] = value
}

func (m Metadata) Get(key string) []byte {
	if m == nil {
		return nil
	}

	return m[key]
}
