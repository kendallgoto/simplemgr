package smp

import (
	"encoding/base64"
	"fmt"
	"slices"

	"github.com/sigurn/crc16"
)

// Framing selects how an SMP message is delimited on the wire.
type Framing uint8

const (
	// PartialFrames is the typical SMP serial behavior: a message is sent as an
	// initial frame (0x06 0x09) optionally followed by partial frames
	// (0x04 0x14), fragmenting the packet so each line stays within a maximum
	// length. It is the default.
	PartialFrames Framing = iota
	// SingleFrames always emits exactly one initial-final frame (0x06 0x09) and
	// never fragments, regardless of packet size.
	SingleFrames
	// Unframed sends the raw smp.Message bytes with no length prefix, CRC, base64,
	// or markers. It is used for transports such as BLE where the GATT layer handles
	// framing.
	Unframed
)

// DefaultMaxFrameLen is the maximum on-wire length of a single frame line
// It matches the conventional SMP serial line cap.
const DefaultMaxFrameLen = 127

var (
	initialMarker = []byte{0x06, 0x09}
	partialMarker = []byte{0x04, 0x14}
)

func wrapFrame(m0, m1 byte, raw []byte) []byte {
	enc := base64.StdEncoding.EncodeToString(raw)
	frame := make([]byte, 2+len(enc)+1)
	frame[0] = m0
	frame[1] = m1
	copy(frame[2:], enc)
	frame[len(frame)-1] = '\n'
	return frame
}

func fragmentSize(maxFrame int) int {
	if maxFrame <= 0 {
		maxFrame = DefaultMaxFrameLen
	}
	groups := (maxFrame - 3) / 4 // 2 marker bytes + 1 newline of overhead
	if groups < 1 {
		groups = 1
	}
	return groups * 3
}

// EncodeFrames serializes msg into the bytes to transmit for the given framing
// mode. For PartialFrames, maxFrame bounds each frame line's on-wire length.
// The packet is split into an initial frame followed by as many partial frames as
// needed, the last of which carries the trailing CRC. SingleFrames always emits one
// frame and Unframed emits the raw message bytes.
func EncodeFrames(msg *Message, mode Framing, maxFrame int) ([]byte, error) {
	if mode == Unframed {
		return msg.Bytes_()
	}

	packet, err := msg.packet()
	if err != nil {
		return nil, err
	}

	frag := fragmentSize(maxFrame)
	if mode == SingleFrames || len(packet) <= frag {
		return wrapFrame(initialMarker[0], initialMarker[1], packet), nil
	}

	// PartialFrames, packet too large for one line: emit an initial frame
	// followed by partial frames. Because frag is a multiple of 3, every frame
	// except the last is offset- and length-aligned to a base64 quantum.
	var out []byte
	for off := 0; off < len(packet); off += frag {
		end := min(off+frag, len(packet))
		m0, m1 := partialMarker[0], partialMarker[1]
		if off == 0 {
			m0, m1 = initialMarker[0], initialMarker[1]
		}
		out = append(out, wrapFrame(m0, m1, packet[off:end])...)
	}
	return out, nil
}

// ParsePacket validates and decodes a reassembled packet payload
// validating the CRC
func ParsePacket(bodyCRC []byte) (*Message, error) {
	if len(bodyCRC) < 2 {
		return nil, fmt.Errorf("frame too short: %d bytes", len(bodyCRC))
	}
	body := bodyCRC[:len(bodyCRC)-2]
	w := crc16.New(crcTable)
	if _, err := w.Write(body); err != nil {
		return nil, fmt.Errorf("while computing crc: %w", err)
	}
	if !slices.Equal(w.Sum(nil), bodyCRC[len(bodyCRC)-2:]) {
		return nil, fmt.Errorf("unexpected crc")
	}
	return NewMessage(body)
}
