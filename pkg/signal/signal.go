package signal

import (
	"context"
	"fmt"
	"os"
	"os/signal"
)

type InterruptedError string

func (e InterruptedError) Error() string {
	return fmt.Sprintf("interrupted by signal %s", string(e))
}

// Handle runs until:
// - The passed context is done.
// - One of the specified signals is received.
// In the later case it returns an InterruptedError.
func Handle(ctx context.Context, signals ...os.Signal) error {
	sigCh := make(chan os.Signal, 1)

	signal.Notify(sigCh, signals...)

	select {
	case sig := <-sigCh:
		return InterruptedError(sig.String())

	case <-ctx.Done():
		return nil
	}
}
