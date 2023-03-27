package entity

import (
	"sync"

	"github.com/vmihailenco/msgpack/v5"
)

type Result struct {
	Format Format
	Err    error
	Files  []File
}

type Results struct {
	mu      sync.RWMutex
	results []Result
}

func (r *Results) MarshalMsgpack() ([]byte, error) {
	return msgpack.Marshal(r.results)
}

func (r *Results) UnmarshalMsgpack(b []byte) error {
	return msgpack.Unmarshal(b, &r.results)
}

func (r *Results) Add(result Result) {
	r.mu.Lock()
	results := r.results
	results = append(results, result)
	r.results = results
	r.mu.Unlock()
}

func (r *Results) Results() []Result {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.results
}
