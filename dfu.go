package simplemgr

import (
	"context"
	"fmt"
	"io"

	goutil "github.com/kendallgoto/goutil"
	"github.com/kendallgoto/simplemgr/pkg/smp"
)

// Amount to upload in a single SMP message
const DefaultDFUChunkSize = 128

// TODO: assumed overhead reserved for SMP framing / CBOR keys; this is pretty generous though
const dfuChunkOverhead = 128

// DFUConfig configures a DFU upload session created with Port.NewDFU.
type DFUConfig struct {
	// Length is the total size of the image in bytes.
	Length uint32
	// Hash is the SHA-256 of the complete image.
	Hash []byte
	// Image selects the target image number; nil uploads to the default slot.
	Image *uint32
	// ChunkSize is the number of image bytes sent per UploadImage request. When
	// it is <= 0, NewDFU queries the device's MCUmgr parameters and picks a size
	// that fits its buffer, falling back to DefaultDFUChunkSize.
	ChunkSize int
	// Confirm marks the uploaded image as permanent on Close. Takes precedence
	// over Test if both are set.
	Confirm bool
	// Test marks the uploaded image for a single test boot on Close.
	Test bool
	// Reset reboots the device on Close, after any Test/Confirm step, so a
	// pending image is swapped in by the bootloader.
	Reset bool
}

// DFU is a streaming firmware upload session. Image data written to it is split
// into UploadImage requests automatically, and Close finalizes the transfer by
// optionally testing/confirming the image and resetting the device. DFU implements
// io.WriteCloser so an image can be streamed in with io.Copy
type DFU struct {
	ctx       context.Context
	port      *Port
	cfg       DFUConfig
	chunkSize int

	buf     []byte // data written but not yet acknowledged by the device
	offset  uint32 // device's next-expected offset; image offset of buf[0]
	written uint32 // total bytes accepted from Write
	closed  bool
}

// NewDFU starts a DFU upload session on the port.
func (p *Port) NewDFU(ctx context.Context, cfg DFUConfig) (*DFU, error) {
	if cfg.Length == 0 {
		return nil, fmt.Errorf("length is required")
	}
	chunk := cfg.ChunkSize
	if chunk <= 0 {
		chunk = DefaultDFUChunkSize
		// Prefer the device's advertised buffer size when available so we send
		// the largest chunks it can accept; ignore errors and fall back.
		if params, err := p.GetMcuMgrParameters(ctx); err == nil && params.BufSize > dfuChunkOverhead {
			chunk = int(params.BufSize) - dfuChunkOverhead
		}
	}
	return &DFU{
		ctx:       ctx,
		port:      p,
		cfg:       cfg,
		chunkSize: chunk,
	}, nil
}

// One call helper for NewDFU: provide the reader and length and it will perform all of the
// copy and install steps.
func (p *Port) UploadImageFrom(ctx context.Context, r io.Reader, length uint32, cfg DFUConfig) error {
	cfg.Length = length
	dfu, err := p.NewDFU(ctx, cfg)
	if err != nil {
		return err
	}
	if _, err := io.Copy(dfu, r); err != nil {
		return err
	}
	return dfu.Close()
}

// ChunkSize reports the per-request data size the session is using.
func (d *DFU) ChunkSize() int { return d.chunkSize }

// Uploaded reports the number of image bytes the device has acknowledged so far.
func (d *DFU) Uploaded() uint32 { return d.offset }

// Write buffers image data and flushes whole chunks to the device. It always
// reports the full slice as consumed unless an upload fails.
func (d *DFU) Write(p []byte) (int, error) {
	if d.closed {
		return 0, fmt.Errorf("write after close")
	}
	if d.written+uint32(len(p)) > d.cfg.Length {
		return 0, fmt.Errorf("wrote more than the declared image length of %d bytes", d.cfg.Length)
	}
	d.buf = append(d.buf, p...)
	d.written += uint32(len(p))
	for len(d.buf) >= d.chunkSize {
		if err := d.flush(); err != nil {
			return 0, err
		}
	}
	return len(p), nil
}

// flush sends a single chunk (up to chunkSize bytes) from the head of the
// buffer and advances by however much the device acknowledges.
func (d *DFU) flush() error {
	if len(d.buf) == 0 {
		return nil
	}
	n := min(d.chunkSize, len(d.buf))
	req := &smp.ImageUploadRequest{
		Offset: d.offset,
		Data:   d.buf[:n],
	}
	// The first request (offset 0) carries the total length, optional image
	// hash, and target image number.
	if d.offset == 0 {
		req.Length = goutil.Ptr(d.cfg.Length)
		if len(d.cfg.Hash) > 0 {
			req.Hash = d.cfg.Hash
		}
		req.Image = d.cfg.Image
	}
	resp, err := d.port.UploadImage(d.ctx, req)
	if err != nil {
		return fmt.Errorf("failed upload at offset %d: %w", d.offset, err)
	}
	// The device reports the next offset it expects. It may accept fewer bytes
	// than we sent, but it must make forward progress and cannot ack data we
	// have not sent.
	if resp.Offset <= d.offset || resp.Offset > d.offset+uint32(n) {
		return fmt.Errorf("device returned invalid offset %d (was at %d, sent %d)", resp.Offset, d.offset, n)
	}
	d.buf = d.buf[resp.Offset-d.offset:]
	d.offset = resp.Offset
	return nil
}

// Close flushes any remaining data, verifies the whole image was uploaded, then
// applies the configured Test/Confirm and Reset steps. It is safe to call once.
func (d *DFU) Close() error {
	if d.closed {
		return nil
	}
	d.closed = true

	for len(d.buf) > 0 {
		if err := d.flush(); err != nil {
			return err
		}
	}
	if d.offset != d.cfg.Length {
		return fmt.Errorf("dfu: upload incomplete, device acknowledged %d of %d bytes", d.offset, d.cfg.Length)
	}

	if d.cfg.Confirm || d.cfg.Test {
		// Confirm makes the image permanent; otherwise mark it for a single test
		// boot. The image hash identifies which slot to act on.
		if _, err := d.port.SetImageState(d.ctx, d.cfg.Hash, d.cfg.Confirm); err != nil {
			return fmt.Errorf("dfu: setting image state: %w", err)
		}
	}
	if d.cfg.Reset {
		if _, err := d.port.Reset(d.ctx, 0); err != nil {
			return fmt.Errorf("dfu: resetting device: %w", err)
		}
	}
	return nil
}
