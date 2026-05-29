package smp

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"slices"

	"github.com/fxamacker/cbor/v2"
	"github.com/kaitai-io/kaitai_struct_go_runtime/kaitai"
	"github.com/sigurn/crc16"
)

var crcTable = crc16.MakeTable(crc16.CRC16_XMODEM)

var (
	ErrUnknown      = errors.New("unknown error")
	ErrBadInput     = errors.New("bad input")
	ErrTimeout      = errors.New("timed out")
	ErrNotFound     = errors.New("not found")
	ErrNotSupported = errors.New("not supported")
	ErrBusy         = errors.New("system busy")
)

var errorByCode = map[Errors]error{
	ErrorsUnknown: ErrUnknown,
	ErrorsInval:   ErrBadInput,
	ErrorsTimeout: ErrTimeout,
	ErrorsNoEnt:   ErrNotFound,
	ErrorsNotSup:  ErrNotSupported,
	ErrorsBusy:    ErrBusy,
}

func errorForCode(code Errors) error {
	if err, ok := errorByCode[code]; ok {
		return err
	}
	return ErrUnknown
}

type groupPayloadBinding struct {
	Structure reflect.Type
	Command   reflect.Type
}

var (
	groupBindings = map[Group]groupPayloadBinding{
		GroupOs: {
			Structure: reflect.TypeFor[Os](),
			Command:   reflect.TypeFor[OsCommand](),
		},
		GroupImage: {
			Structure: reflect.TypeFor[Image](),
			Command:   reflect.TypeFor[ImageCommand](),
		},
		GroupStat: {
			Structure: reflect.TypeFor[Stat](),
			Command:   reflect.TypeFor[StatCommand](),
		},
		GroupSettings: {
			Structure: reflect.TypeFor[Settings](),
			Command:   reflect.TypeFor[SettingsCommand](),
		},
		GroupLog: {
			Structure: reflect.TypeFor[Generic](),
			Command:   reflect.TypeFor[uint8](),
		},
		GroupCrash: {
			Structure: reflect.TypeFor[Generic](),
			Command:   reflect.TypeFor[uint8](),
		},
		GroupSplit: {
			Structure: reflect.TypeFor[Generic](),
			Command:   reflect.TypeFor[uint8](),
		},
		GroupRun: {
			Structure: reflect.TypeFor[Generic](),
			Command:   reflect.TypeFor[uint8](),
		},
		GroupFs: {
			Structure: reflect.TypeFor[Fs](),
			Command:   reflect.TypeFor[FsCommand](),
		},
		GroupShell: {
			Structure: reflect.TypeFor[Shell](),
			Command:   reflect.TypeFor[ShellCommand](),
		},
		GroupZephyr: {
			Structure: reflect.TypeFor[Zephyr](),
			Command:   reflect.TypeFor[ZephyrCommand](),
		},
	}
)

type intConstraint interface {
	~int
}

// NewRequest builds an SMP message for the given operation, group, and command, with a CBOR-encoded payload.
func NewRequest[C intConstraint](op Operation, group Group, command C, payload any) (*Message, error) {
	reqBytes, err := cbor.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("while marshalling to cbor: %w", err)
	}

	binding, ok := groupBindings[group]
	if !ok {
		return nil, fmt.Errorf("no bindings for mcumgr group %d", group)
	}
	structure := reflect.New(binding.Structure)
	structure.Elem().FieldByName("Command").Set(reflect.ValueOf(command))
	structure.Elem().FieldByName("Payload").Set(reflect.ValueOf(reqBytes))
	return &Message{
		Ver:    1,
		Op:     op,
		Group:  group,
		Length: uint16(len(reqBytes)),
		Body:   structure.Interface().(kaitai.Struct),
	}, nil

}

// DecodeMessageBody parses an SMP message and unmarshals the CBOR payload into the provided structure.
func DecodeMessageBody(msg *Message, out any) error {
	if msg == nil || msg.Body == nil {
		return fmt.Errorf("nil message body")
	}
	outV := reflect.ValueOf(out)
	if outV.Kind() != reflect.Pointer || outV.IsNil() {
		return fmt.Errorf("decode target must be a non-nil pointer")
	}
	bodyV := reflect.ValueOf(msg.Body)
	if bodyV.Kind() != reflect.Pointer || bodyV.IsNil() {
		return fmt.Errorf("unable to decode msg type")
	}
	payloadF := bodyV.Elem().FieldByName("Payload")
	if !payloadF.IsValid() || !payloadF.CanInterface() {
		return fmt.Errorf("unable to decode msg type")
	}
	payload, ok := payloadF.Interface().([]byte)
	if !ok {
		return fmt.Errorf("unexpected payload type")
	}
	if err := cbor.Unmarshal(payload, out); err != nil {
		return fmt.Errorf("while decoding cbor payload: %w", err)
	}

	elem := outV.Elem()
	if elem.Kind() != reflect.Struct {
		return nil
	}
	// SMP v1 / non-group return code.
	if errCode := elem.FieldByName("ErrorCode"); errCode.IsValid() && errCode.Kind() == reflect.Pointer && !errCode.IsNil() {
		if code := Errors(errCode.Elem().Uint()); code != ErrorsOk {
			return errorForCode(code)
		}
	}
	// TODO: SMP v2 reports group-specific error codes in the "err" map (a group
	// id plus a group-relative rc). Decoding those correctly requires per-group
	// error tables; for now, surface any structured v2 error as ErrUnknown.
	if errResponse := elem.FieldByName("Error"); errResponse.IsValid() && errResponse.Kind() == reflect.Pointer && !errResponse.IsNil() {
		return ErrUnknown
	}
	return nil
}

// MarshalBinary encodes the message into its framed UART form: a base64-wrapped
// body with a length prefix and CRC suffix. It implements encoding.BinaryMarshaler.
func (msg *Message) MarshalBinary() ([]byte, error) {
	body, err := msg.Bytes_()
	if err != nil {
		return nil, fmt.Errorf("while packing smp message: %v", err)
	}

	// Build the packet: 2-byte length prefix + body + 2-byte CRC (XMODEM).
	packet := make([]byte, len(body)+4)
	binary.BigEndian.PutUint16(packet, uint16(len(body)+2))
	copy(packet[2:], body)
	w := crc16.New(crcTable)
	if _, err = w.Write(body); err != nil {
		return nil, fmt.Errorf("while computing crc: %w", err)
	}
	copy(packet[len(body)+2:], w.Sum(nil))

	// The packet is base64 wrapped, prefixed with the 0x0609 non-fragmented
	// start-of-frame marker and terminated by a newline.
	encoded := base64.StdEncoding.EncodeToString(packet)
	frame := make([]byte, 2+base64.StdEncoding.EncodedLen(len(packet))+1)
	frame[0] = 0x06
	frame[1] = 0x09
	copy(frame[2:], encoded)
	frame[len(frame)-1] = '\n'
	return frame, nil
}

// UnmarshalBinary decodes a framed UART line into the message. It strips any
// framing, base64-decodes the payload, and validates the CRC. It implements
// encoding.BinaryUnmarshaler.
func (msg *Message) UnmarshalBinary(data []byte) error {
	// if we receive data that is framed, drop the framing
	data = bytes.TrimSuffix(data, []byte{'\n'})
	data = bytes.TrimPrefix(data, []byte{0x06, 0x09})

	decoded, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return fmt.Errorf("decode base64: %w", err)
	}
	if len(decoded) < 4 {
		return fmt.Errorf("frame too short: %d bytes", len(decoded))
	}

	// Strip 2-byte length prefix and 2-byte CRC suffix.
	if len(decoded)-2 != int(binary.BigEndian.Uint16(decoded[0:2])) {
		return fmt.Errorf("unexpected length")
	}
	w := crc16.New(crcTable)
	if _, err = w.Write(decoded[2 : len(decoded)-2]); err != nil {
		return fmt.Errorf("while computing crc: %w", err)
	}
	if !slices.Equal(w.Sum(nil), decoded[len(decoded)-2:]) {
		return fmt.Errorf("unexpected crc")
	}
	parsed, err := NewMessage(decoded[2 : len(decoded)-2])
	if err != nil {
		return err
	}
	*msg = *parsed
	return nil
}
