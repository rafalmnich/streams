package streams_test

import (
	"context"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/msales/streams/v3"
	"github.com/stretchr/testify/assert"
)

func TestMessage_Empty(t *testing.T) {
	tests := []struct {
		key   interface{}
		value interface{}
		empty bool
	}{
		{"test", "test", false},
		{nil, "test", false},
		{"test", nil, false},
		{nil, nil, true},
	}

	for _, tt := range tests {
		msg := streams.NewMessage(tt.key, tt.value)

		assert.Equal(t, tt.empty, msg.Empty())
	}
}

func TestNewMessage(t *testing.T) {
	msg := streams.NewMessage("test", "test")

	assert.Equal(t, context.Background(), msg.Ctx)
	assert.Equal(t, "test", msg.Key)
	assert.Equal(t, "test", msg.Value)
}

type ctxKey string

func TestNewMessageWithContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), ctxKey("1"), "2")
	msg := streams.NewMessageWithContext(ctx, "test", "test")

	assert.Equal(t, ctx, msg.Ctx)
	assert.Equal(t, "test", msg.Key)
	assert.Equal(t, "test", msg.Value)
}

func TestMessage_Metadata(t *testing.T) {
	s := new(MockSource)
	m := new(MockMetadata)
	msg := streams.NewMessage("test", "test")

	msg = msg.WithMetadata(s, m)

	src, meta := msg.Metadata()
	assert.Equal(t, s, src)
	assert.Equal(t, m, meta)
}

func TestMessageBuffer(t *testing.T) {
	const N = 10000
	var wg sync.WaitGroup

	buf := streams.NewMessageBuffer(1024)

	wg.Add(2)
	go func() {
		defer wg.Done()

		for i := 0; i < N; i++ {
			buf.Write(streams.NewMessage(nil, i))
			time.Sleep(1 * time.Microsecond)
		}
		buf.Close()
	}()

	read := 0
	go func() {
		defer wg.Done()

		var msgs [100]streams.Message
		i := 0
		for !buf.Done() {
			n := buf.Read(msgs[:])
			for _, msg := range msgs[:n] {
				assert.Equal(t, i, msg.Value)
				i++
				read++
			}
		}
	}()

	wg.Wait()

	assert.Equal(t, N, read)
}

func TestMessageBuffer_SyncRead(t *testing.T) {
	const N = 10000

	buf := streams.NewMessageBuffer(1024)

	go func() {
		for i := 0; i < N; i++ {
			buf.Write(streams.NewMessage(nil, i))
			time.Sleep(1 * time.Microsecond)
		}
		buf.Close()
	}()

	time.Sleep(time.Second)

	read := 0
	var msgs [100]streams.Message
	i := 0
	for !buf.Done() {
		n := buf.Read(msgs[:])
		for _, msg := range msgs[:n] {
			assert.Equal(t, i, msg.Value)
			i++
			read++
		}
	}

	assert.Equal(t, N, read)
}

func TestMessageBuffer_SmallBuffer(t *testing.T) {
	const N = 10000

	buf := streams.NewMessageBuffer(10)

	go func() {
		for i := 0; i < N; i++ {
			buf.Write(streams.NewMessage(nil, i))
			time.Sleep(1 * time.Microsecond)
		}
		buf.Close()
	}()

	time.Sleep(time.Second)

	read := 0
	var msgs [100]streams.Message
	i := 0
	for !buf.Done() {
		n := buf.Read(msgs[:])
		for _, msg := range msgs[:n] {
			assert.Equal(t, i, msg.Value)
			i++
			read++
		}
	}

	assert.Equal(t, N, read)
}

func BenchmarkMessageBuffer_NoPin(b *testing.B) {
	var wg sync.WaitGroup

	buf := streams.NewMessageBuffer(1024)
	msg := streams.NewMessage(nil, "hello")

	b.ReportAllocs()
	b.ResetTimer()

	wg.Add(2)
	go func() {
		for i := 0; i < b.N; i++ {
			buf.Write(msg)
		}
		buf.Close()
		wg.Done()
	}()

	go func() {
		i := 0
		var msgs [100]streams.Message
		for !buf.Done() {
			n := buf.Read(msgs[:])
			for _, msg := range msgs[:n] {
				_ = msg
				i++
			}
		}
		wg.Done()
	}()

	wg.Wait()
}

func BenchmarkMessageBuffer_Pin(b *testing.B) {
	var wg sync.WaitGroup

	buf := streams.NewMessageBuffer(1024)
	msg := streams.NewMessage(nil, "hello")

	b.ReportAllocs()
	b.ResetTimer()

	wg.Add(2)
	go func() {
		runtime.LockOSThread()
		for i := 0; i < b.N; i++ {
			buf.Write(msg)
		}
		buf.Close()
		wg.Done()
	}()

	go func() {
		runtime.LockOSThread()
		i := 0
		var msgs [100]streams.Message
		for !buf.Done() {
			n := buf.Read(msgs[:])
			for _, msg := range msgs[:n] {
				_ = msg
				i++
			}
		}
		wg.Done()
	}()

	wg.Wait()
}
