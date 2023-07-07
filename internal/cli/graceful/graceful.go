package graceful

import (
	"context"
	"errors"
	"net"
	"net/http"
)

// Serve the handler at address
func Serve(ctx context.Context, listener net.Listener, handler http.Handler) error {
	// Create the HTTP server
	server := &http.Server{Addr: listener.Addr().String(), Handler: handler}
	// Make the server shutdownable
	shutdown := shutdown(ctx, server)
	// Serve requests
	if err := server.Serve(listener); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}
	// Handle any errors that occurred while shutting down
	if err := <-shutdown; err != nil {
		if !errors.Is(err, context.Canceled) {
			return err
		}
	}
	return nil
}

// Shutdown the server when the context is canceled
func shutdown(ctx context.Context, server *http.Server) <-chan error {
	shutdown := make(chan error, 1)
	go func() {
		<-ctx.Done()
		// Shutdown the server immediately
		if err := server.Close(); err != nil {
			shutdown <- err
		}
		close(shutdown)
	}()
	return shutdown
}
