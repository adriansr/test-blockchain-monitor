package metrics

import (
	"encoding/json"
	"io"
	"sync"

	"monitor/pkg/model"
)

type Exemplar struct {
	mu     sync.Mutex
	latest model.Block
}

func (e *Exemplar) OnEvent(block model.Block) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.latest = block
	return nil
}

func (e *Exemplar) get() (model.Block, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.latest.Header.Number == nil {
		return e.latest, false
	}
	return e.latest, true
}

func (e *Exemplar) Output(w io.Writer) error {
	b, ok := e.get()
	if !ok {
		return nil
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	return enc.Encode(b)
}
