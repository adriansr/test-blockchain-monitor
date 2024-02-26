package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"monitor/pkg/client"
	"monitor/pkg/metrics"
	"monitor/pkg/signal"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/sync/errgroup"
)

var (
	flagNodeAddress   = flag.String("node", "wss://localhost", "Address to connect to")
	flagServerAddress = flag.String("http-serve", "localhost:8989", "Address for local HTTP server")
	flagTimeout       = flag.Duration("timeout", time.Second*15, "default i/o timeout")
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}
}

func run() error {
	flag.Parse() // Exits on error

	// Create Ethereum client
	cli, err := client.New(client.Opts{
		Address:        *flagNodeAddress,
		ConnectTimeout: *flagTimeout,
	})
	if err != nil {
		return err
	}

	// Create prometheus registry for metrics
	registerer := prometheus.NewRegistry()

	// Create consumers for new blocks
	var (
		// logger logs blocks to stdout
		logger = metrics.NewLogger(os.Stdout)

		// head keeps the latest block observed
		head = &metrics.Exemplar{}
	)

	stats, err := metrics.NewStats(registerer)
	if err != nil {
		return err
	}

	cli.Subscribe(logger)
	cli.Subscribe(head)
	cli.Subscribe(stats)

	// Create HTTP Server

	listener, err := net.Listen("tcp", *flagServerAddress)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()

	// /head endpoint will print the latest block observed (head)
	mux.HandleFunc("/head", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		head.Output(w)
	})

	// /metrics endpoint will output prometheus metrics
	mux.Handle("/metrics", promhttp.HandlerFor(registerer, promhttp.HandlerOpts{}))

	// Run cancellable blockchain consumer, http server and signal handler

	group, ctx := errgroup.WithContext(context.Background())
	group.Go(func() error {
		return cli.Run(ctx)
	})

	group.Go(func() error {
		errChan := make(chan error, 1)
		go func() {

			server := http.Server{
				Handler:           mux,
				ReadTimeout:       *flagTimeout,
				ReadHeaderTimeout: *flagTimeout,
				WriteTimeout:      *flagTimeout,
				IdleTimeout:       *flagTimeout,
			}
			errChan <- server.Serve(listener)
		}()

		select {
		case err := <-errChan:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	})

	group.Go(func() error {
		return signal.Handle(ctx, os.Kill, os.Interrupt)
	})
	return group.Wait()
}
