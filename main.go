package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	port int
)

func init() {
	flag.IntVar(&port, "port", 8080, "Specify the port to listen to")
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	go func() { // catch signal and invoke graceful termination
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop

		log.Printf("[DEBUG] interrupt signal")
		cancel()
	}()

	v := NewVisitor()
	s := NewServer(port, 100, v)
	s.Run(ctx)

}
