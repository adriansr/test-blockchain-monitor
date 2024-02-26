package metrics

import (
	"fmt"
	"io"
	"time"

	"monitor/pkg/model"
)

type Logger struct {
	w io.Writer
}

func NewLogger(w io.Writer) Logger {
	return Logger{
		w: w,
	}
}

func (l Logger) OnEvent(block model.Block) error {
	fmt.Fprintf(l.w, "%s block[%d] hash=%s parent=%s coinbase=%s logs=%d\n",
		time.Unix(int64(block.Header.Time), 0).Format(time.RFC3339),
		block.Header.Number.Uint64(),
		block.Header.Hash(),
		block.Header.ParentHash.String(),
		block.Header.Coinbase.String(),
		len(block.Logs),
	)
	return nil
}
