package signals

import (
	"os"
	"os/signal"
	"syscall"
)

var onlyOneSignalHandler = make(chan struct{})

var shutdownSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}

// SetupNotifySignalHandler registers SIGTERM and SIGINT.
// When caught the first signal, the stop channel will be closed.
// When a second signal caught, the program is terminated with exit code 1.
func SetupNotifySignalHandler() (stopCh <-chan struct{}) {
	close(onlyOneSignalHandler) // panics when called twice

	stop := make(chan struct{})
	signalCh := make(chan os.Signal, 2)
	signal.Notify(signalCh, shutdownSignals...)
	go func() {
		<-signalCh
		close(stop)
		<-signalCh
		os.Exit(1)
	}()

	return stop
}
