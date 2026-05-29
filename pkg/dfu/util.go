package dfu

import (
	"fmt"
	"io"

	"github.com/kaitai-io/kaitai_struct_go_runtime/kaitai"
)

// GetHashFromTLV reads a MCUboot DFU binary and returns the encoded SHA256 hash value.
func GetHashFromTLV(file io.ReadSeeker) ([]byte, error) {
	pkg := &Package{}
	err := pkg.Read(kaitai.NewStream(file), nil, pkg)
	if err != nil {
		return nil, err
	}
	for _, tlv := range pkg.Tlvs.TlvContainer.Tlv {
		if tlv.Type == TlvTypesSha256 {
			return tlv.TlvInner, nil
		}
	}
	return nil, fmt.Errorf("unable to find hash TLV")
}
