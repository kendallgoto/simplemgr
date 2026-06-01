// Package simplemgr implements the host interface for SMP commands as seen in MCUmgr / MCUboot serial recovery.
// Its primary purpose is to provide recovery flashing functions to a UART connected device, however it can also
// be used to send SMP commands to a running device via Bluetooth / runtime serial interface.
//
// MCUmgr specification is based on Zephyr documentation https://docs.zephyrproject.org/3.7.0/services/device_mgmt/mcumgr.html
package simplemgr

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"slices"
	"sync"
	"time"

	"github.com/kendallgoto/simplemgr/pkg/smp"
)

// DefaultTimeout is the maximum time to wait for a single operation before it is aborted.
const DefaultTimeout = 3 * time.Second

const readPollInterval = 50 * time.Millisecond

// Framing selects the UART framing mode for a Port. See the smp package for the
// behavior of each mode.
type Framing = smp.Framing

// Framing modes to use for message packing - different transports require different framing modes
// see https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_transport.html
const (
	PartialFrames = smp.PartialFrames
	SingleFrames  = smp.SingleFrames
	Unframed      = smp.Unframed
)

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
	// Framing selects the type of frame to wrap around the messages, defaults to PartialFrames.
	Framing Framing
	// MaxFrameLen sets the max size for each frame, defaulting to SMP standard 127 bytes
	MaxFrameLen int

	seq uint8
	rx  []byte // bytes read from Device but not yet consumed by the framer
}

// New creates a Port for the given device using DefaultTimeout and default framing
func New(device io.ReadWriteCloser) *Port {
	return &Port{
		Device:      device,
		Timeout:     DefaultTimeout,
		Framing:     PartialFrames,
		MaxFrameLen: smp.DefaultMaxFrameLen,
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

	wire, err := smp.EncodeFrames(msg, p.Framing, p.MaxFrameLen)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if _, err = p.Device.Write(wire); err != nil {
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
		resp, err := p.readMessage(ctx)
		resultCh <- result{msg: resp, err: err}
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

func (p *Port) readMessage(ctx context.Context) (*smp.Message, error) {
	if p.Framing == Unframed {
		return p.readUnframed(ctx)
	}

	initial, content, err := p.readLine(ctx)
	if err != nil {
		return nil, err
	}
	if !initial {
		return nil, fmt.Errorf("expected an initial frame, got a partial frame")
	}
	raw, err := base64.StdEncoding.DecodeString(string(content))
	if err != nil {
		return nil, fmt.Errorf("decode base64: %w", err)
	}
	if len(raw) < 2 {
		return nil, fmt.Errorf("initial frame too short: %d bytes", len(raw))
	}

	// The initial frame leads with the total length of the body and CRC that
	// follow; keep reading partial frames until that many bytes are gathered.
	total := int(binary.BigEndian.Uint16(raw[:2]))
	acc := slices.Clone(raw[2:])
	for len(acc) < total {
		more, content, err := p.readLine(ctx)
		if err != nil {
			return nil, err
		}
		if more {
			return nil, fmt.Errorf("unexpected initial frame mid-packet")
		}
		dec, err := base64.StdEncoding.DecodeString(string(content))
		if err != nil {
			return nil, fmt.Errorf("decode base64: %w", err)
		}
		acc = append(acc, dec...)
	}
	return smp.ParsePacket(acc[:total])
}

func (p *Port) readUnframed(ctx context.Context) (*smp.Message, error) {
	var buf [256]byte
	for {
		if len(p.rx) >= 8 {
			if msg, err := smp.NewMessage(p.rx); err == nil {
				p.rx = nil
				return msg, nil
			}
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		n, err := p.Device.Read(buf[:])
		if err != nil {
			return nil, err
		}
		p.rx = append(p.rx, buf[:n]...)
	}
}

func (p *Port) readLine(ctx context.Context) (initial bool, content []byte, err error) {
	var buf [256]byte
	for {
		if init, body, ok := p.takeLine(); ok {
			return init, body, nil
		}
		if err := ctx.Err(); err != nil {
			return false, nil, err
		}
		n, err := p.Device.Read(buf[:])
		if err != nil {
			return false, nil, err
		}
		p.rx = append(p.rx, buf[:n]...)
	}
}

func (p *Port) takeLine() (initial bool, content []byte, ok bool) {
	for i := 0; i+1 < len(p.rx); i++ {
		init := p.rx[i] == 0x06 && p.rx[i+1] == 0x09
		part := p.rx[i] == 0x04 && p.rx[i+1] == 0x14
		if !init && !part {
			continue
		}
		rest := p.rx[i+2:]
		nl := bytes.IndexByte(rest, '\n')
		if nl < 0 {
			// Marker seen but the line is incomplete; drop preceding junk and
			// wait for the rest of the line.
			p.rx = slices.Clone(p.rx[i:])
			return false, nil, false
		}
		content = slices.Clone(rest[:nl])
		p.rx = slices.Clone(rest[nl+1:])
		return init, content, true
	}
	// No marker found; retain only a trailing byte that might begin a marker
	// split across reads.
	if len(p.rx) > 1 {
		p.rx = slices.Clone(p.rx[len(p.rx)-1:])
	}
	return false, nil, false
}
