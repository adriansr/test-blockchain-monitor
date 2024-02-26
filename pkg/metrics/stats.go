package metrics

import (
	"monitor/pkg/model"

	"github.com/prometheus/client_golang/prometheus"
)

type Stats struct {
	blkCount prometheus.Counter
	txCount  prometheus.Counter
}

func NewStats(reg prometheus.Registerer) (*Stats, error) {
	var st Stats
	st.txCount = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "blockchain",
		Subsystem: "transaction",
		Name:      "count",
		Help:      "number of transactions processed",
	})
	if err := reg.Register(st.txCount); err != nil {
		return nil, err
	}

	st.blkCount = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "blockchain",
		Subsystem: "block",
		Name:      "count",
		Help:      "number of blocks processed",
	})
	if err := reg.Register(st.blkCount); err != nil {
		return nil, err
	}

	return &st, nil
}

func (s *Stats) OnEvent(block model.Block) error {
	s.blkCount.Inc()

	txSeen := make(map[string]struct{}, len(block.Logs))
	txSeen[block.Header.TxHash.String()] = struct{}{}
	for _, log := range block.Logs {
		txSeen[log.TxHash.String()] = struct{}{}
	}

	s.txCount.Add(float64(len(txSeen)))
	return nil
}
