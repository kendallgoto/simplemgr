package simplemgr

import (
	"context"

	goutil "github.com/kendallgoto/goutil"
	"github.com/kendallgoto/simplemgr/pkg/smp"
)

// GetImageState returns a list of all the installed images, their version, status, and other misc information.
func (p *Port) GetImageState(ctx context.Context) (*smp.ImageStateResponse, error) {
	msg, err := smp.NewRequest(smp.OperationRead, smp.GroupImage, smp.ImageCommandState, &smp.ImageStateRequest{})
	if err != nil {
		return nil, err
	}
	resp, err := p.writeAndRead(ctx, msg)
	if err != nil {
		return nil, err
	}
	out := &smp.ImageStateResponse{}
	if err := smp.DecodeMessageBody(resp, out); err != nil {
		return nil, err
	}
	return out, nil
}

// SetImageState is used to confirm/test an image that is pending.
func (p *Port) SetImageState(ctx context.Context, hash []byte, confirm bool) (*smp.ImageStateResponse, error) {
	msg, err := smp.NewRequest(smp.OperationWrite, smp.GroupImage, smp.ImageCommandState, &smp.ImageStateRequest{
		Hash:    hash,
		Confirm: confirm,
	})
	if err != nil {
		return nil, err
	}
	resp, err := p.writeAndRead(ctx, msg)
	if err != nil {
		return nil, err
	}
	out := &smp.ImageStateResponse{}
	if err := smp.DecodeMessageBody(resp, out); err != nil {
		return nil, err
	}
	return out, nil
}

// UploadImage is used to upload a chunk of an image to a given slot. This is the main function needed
// to perform DFU, by calling UploadImage multiple times for the image fragments being uploaded.
func (p *Port) UploadImage(ctx context.Context, req *smp.ImageUploadRequest) (*smp.ImageUploadResponse, error) {
	msg, err := smp.NewRequest(smp.OperationWrite, smp.GroupImage, smp.ImageCommandUpload, req)
	if err != nil {
		return nil, err
	}
	resp, err := p.writeAndRead(ctx, msg)
	if err != nil {
		return nil, err
	}
	out := &smp.ImageUploadResponse{}
	if err := smp.DecodeMessageBody(resp, out); err != nil {
		return nil, err
	}
	return out, nil
}

// EraseImage tells the device to erase a firmware slot.
func (p *Port) EraseImage(ctx context.Context, slot uint8) (*smp.ImageEraseResponse, error) {
	msg, err := smp.NewRequest(smp.OperationWrite, smp.GroupImage, smp.ImageCommandErase, &smp.ImageEraseRequest{
		Slot: goutil.Ptr(slot),
	})
	if err != nil {
		return nil, err
	}
	resp, err := p.writeAndRead(ctx, msg)
	if err != nil {
		return nil, err
	}
	out := &smp.ImageEraseResponse{}
	if err := smp.DecodeMessageBody(resp, out); err != nil {
		return nil, err
	}
	return out, nil
}
