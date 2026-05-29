// Package simplemgr implements the host interface for SMP commands as seen in MCUmgr / MCUboot serial recovery.
// Its primary purpose is to provide recovery flashing functions to a UART connected device, however it can also
// be used to send SMP commands to a running device via Bluetooth / runtime serial interface.
//
// MCUmgr specification is based on Zephyr documentation https://docs.zephyrproject.org/3.7.0/services/device_mgmt/mcumgr.html
package simplemgr

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/kendallgoto/simplemgr/pkg/smp"
)

// DefaultTimeout is the maximum time to wait for a single operation before it is aborted.
const DefaultTimeout = 3 * time.Second

const readPollInterval = 50 * time.Millisecond

// readDeadliner is implemented by devices that support a controllable read
// timeout, allowing reads to avoid blocking indefinitely.
type readDeadliner interface {
	SetReadTimeout(t time.Duration) error
}

// Port holds the state for an SMP session over a single device. A Port is not
// safe for concurrent use; it expects one in-flight request at a time.
type Port struct {
	// Device is the underlying transport, typically a serial port.
	Device io.ReadWriteCloser
	// Timeout bounds each request/response exchange; zero disables it.
	Timeout time.Duration

	seq uint8
}

// New creates a Port for the given device using DefaultTimeout.
func New(device io.ReadWriteCloser) *Port {
	return &Port{
		Device:  device,
		Timeout: DefaultTimeout,
	}
}

func (p *Port) writeAndRead(ctx context.Context, msg *smp.Message) (*smp.Message, error) {
	if p.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, p.Timeout)
		defer cancel()
	}

	msg.Seq = p.seq
	p.seq++

	framed, err := msg.MarshalBinary()
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	_, err = p.Device.Write(framed)
	if err != nil {
		return nil, err
	}

	if rd, ok := p.Device.(readDeadliner); ok {
		if err := rd.SetReadTimeout(readPollInterval); err != nil {
			return nil, err
		}
	}

	type result struct {
		msg *smp.Message
		err error
	}
	resultCh := make(chan result, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		encoded, err := p.readFrame(ctx)
		if err != nil {
			resultCh <- result{err: err}
			return
		}
		resp := &smp.Message{}
		if err := resp.UnmarshalBinary(encoded); err != nil {
			resultCh <- result{err: fmt.Errorf("while reading response: %w", err)}
			return
		}
		resultCh <- result{msg: resp}
	}()

	select {
	case <-ctx.Done():
		wg.Wait()
		return nil, ctx.Err()
	case res := <-resultCh:
		if res.err == nil && res.msg != nil && res.msg.Seq != msg.Seq {
			return nil, fmt.Errorf("response seq %d did not match request seq %d", res.msg.Seq, msg.Seq)
		}
		return res.msg, res.err
	}
}

func (p *Port) readFrame(ctx context.Context) ([]byte, error) {
	var (
		body    []byte
		buf     [256]byte
		prev    byte
		started bool
	)
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		n, err := p.Device.Read(buf[:])
		if err != nil {
			return nil, err
		}

		for _, b := range buf[:n] {
			if !started {
				if prev == 0x06 && b == 0x09 {
					started = true
				}
				prev = b
				continue
			}
			if b == '\n' {
				return body, nil
			}
			body = append(body, b)
		}
	}
}
