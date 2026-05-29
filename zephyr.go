package simplemgr

import (
	"context"

	"github.com/kendallgoto/simplemgr/pkg/smp"
)

// EraseStorage tells the device to perform its Zephyr-defined erase functionality, for example
// erasing persisted NVS settings.
func (p *Port) EraseStorage(ctx context.Context) (*smp.ZephyrEraseResponse, error) {
	msg, err := smp.NewRequest(smp.OperationWrite, smp.GroupZephyr, smp.ZephyrCommandErase, &smp.ZephyrEraseRequest{})
	if err != nil {
		return nil, err
	}
	resp, err := p.writeAndRead(ctx, msg)
	if err != nil {
		return nil, err
	}
	out := &smp.ZephyrEraseResponse{}
	if err := smp.DecodeMessageBody(resp, out); err != nil {
		return nil, err
	}
	return out, nil
}
