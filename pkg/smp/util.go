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

// packet builds the framed packet payload: a 2-byte big-endian length prefix
// (covering the body and CRC that follow), the message body, and a 2-byte
// CRC16-XMODEM over that body.
func (msg *Message) packet() ([]byte, error) {
	body, err := msg.Bytes_()
	if err != nil {
		return nil, fmt.Errorf("while packing smp message: %v", err)
	}

	packet := make([]byte, len(body)+4)
	binary.BigEndian.PutUint16(packet, uint16(len(body)+2))
	copy(packet[2:], body)
	w := crc16.New(crcTable)
	if _, err = w.Write(body); err != nil {
		return nil, fmt.Errorf("while computing crc: %w", err)
	}
	copy(packet[len(body)+2:], w.Sum(nil))
	return packet, nil
}

// MarshalBinary encodes the message into a single framed UART line: a
// base64-wrapped packet behind the 0x0609 initial-final marker. It implements
// encoding.BinaryMarshaler and is equivalent to EncodeFrames in SingleFrames mode.
func (msg *Message) MarshalBinary() ([]byte, error) {
	packet, err := msg.packet()
	if err != nil {
		return nil, err
	}
	return wrapFrame(0x06, 0x09, packet), nil
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
