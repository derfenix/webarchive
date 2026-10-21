package badgers3

import (
	"github.com/vmihailenco/msgpack/v5"
)

func marshal(v interface{}) ([]byte, error) {
	return msgpack.Marshal(v)
}

func unmarshal(b []byte, v interface{}) error { //nolint:unused // will use later
	return msgpack.Unmarshal(b, v)
}
