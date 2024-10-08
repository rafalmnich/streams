package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"

	"github.com/rafalmnich/streams/v6"
)

func main() {
	builder := streams.NewStreamBuilder()

	stream1 := builder.Source("rand1-source", newRandIntSource()).
		FilterFunc("filter1", lowNumberFilter)

	builder.Source("rand2-source", newRandIntSource()).
		FilterFunc("filter2", highNumberFilter).
		MapFunc("add-hundedred-mapper", addHundredMapper).
		Merge("merge", stream1).
		Print("print")

	tp, _ := builder.Build()
	task := streams.NewTask(tp)
	task.OnError(func(err error) {
		log.Fatal(err.Error())
	})
	task.Start(context.Background())
	defer task.Close()

	// Wait for SIGTERM
	waitForShutdown()
}

type randIntSource struct {
	rand *rand.Rand
}

func newRandIntSource() streams.Source {
	return &randIntSource{
		rand: rand.New(rand.NewSource(1234)),
	}
}

func (s *randIntSource) Consume() (streams.Message, error) {
	return streams.NewMessage(nil, s.rand.Intn(100)), nil
}

func (s *randIntSource) Commit(v interface{}) error {
	return nil
}

func (s *randIntSource) Close() error {
	return nil
}

func lowNumberFilter(msg streams.Message) (bool, error) {
	num := msg.Value.(int)

	return num < 50, nil
}

func highNumberFilter(msg streams.Message) (bool, error) {
	num := msg.Value.(int)

	return num >= 50, nil
}

func addHundredMapper(msg streams.Message) (streams.Message, error) {
	num := msg.Value.(int)
	msg.Value = num + 100

	return msg, nil
}

// waitForShutdown blocks until a SIGINT or SIGTERM is received.
func waitForShutdown() {
	quit := make(chan os.Signal)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	<-quit
}
