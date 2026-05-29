package simplemgr

import (
	"context"

	goutil "github.com/kendallgoto/goutil"
	"github.com/kendallgoto/simplemgr/pkg/smp"
)

// Echo performs a readback test on the device to echo back the input.
func (p *Port) Echo(ctx context.Context, data string) (*smp.OSEchoResponse, error) {
	msg, err := smp.NewRequest(smp.OperationWrite, smp.GroupOs, smp.OsCommandEcho, &smp.OSEchoRequest{
		Data: data,
	})
	if err != nil {
		return nil, err
	}
	resp, err := p.writeAndRead(ctx, msg)
	if err != nil {
		return nil, err
	}
	out := &smp.OSEchoResponse{}
	if err := smp.DecodeMessageBody(resp, out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetTaskStats retrieves current process tasks and their statistics from the device.
func (p *Port) GetTaskStats(ctx context.Context) (*smp.OSTaskStatsResponse, error) {
	msg, err := smp.NewRequest(smp.OperationRead, smp.GroupOs, smp.OsCommandTaskStats, &smp.OSTaskStatsRequest{})
	if err != nil {
		return nil, err
	}
	resp, err := p.writeAndRead(ctx, msg)
	if err != nil {
		return nil, err
	}
	out := &smp.OSTaskStatsResponse{}
	if err := smp.DecodeMessageBody(resp, out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetDatetime retrieves the current system datetime from the device.
func (p *Port) GetDatetime(ctx context.Context) (*smp.OSDatetimeReadResponse, error) {
	msg, err := smp.NewRequest(smp.OperationRead, smp.GroupOs, smp.OsCommandDatetime, &smp.OSDatetimeReadRequest{})
	if err != nil {
		return nil, err
	}
	resp, err := p.writeAndRead(ctx, msg)
	if err != nil {
		return nil, err
	}
	out := &smp.OSDatetimeReadResponse{}
	if err := smp.DecodeMessageBody(resp, out); err != nil {
		return nil, err
	}
	return out, nil
}

// SetDatetime allows the datetime to be set; it takes a formatted date string per the SMP specification.
func (p *Port) SetDatetime(ctx context.Context, datetime string) (*smp.OSDatetimeWriteResponse, error) {
	msg, err := smp.NewRequest(smp.OperationWrite, smp.GroupOs, smp.OsCommandDatetime, &smp.OSDatetimeWriteRequest{
		Datetime: datetime,
	})
	if err != nil {
		return nil, err
	}
	resp, err := p.writeAndRead(ctx, msg)
	if err != nil {
		return nil, err
	}
	out := &smp.OSDatetimeWriteResponse{}
	if err := smp.DecodeMessageBody(resp, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Reset tells the device to perform a reset. If the device is busy, the reset can be forced.
func (p *Port) Reset(ctx context.Context, force uint8) (*smp.OSResetResponse, error) {
	msg, err := smp.NewRequest(smp.OperationWrite, smp.GroupOs, smp.OsCommandReset, &smp.OSResetRequest{
		Force: goutil.Ptr(force),
	})
	if err != nil {
		return nil, err
	}
	resp, err := p.writeAndRead(ctx, msg)
	if err != nil {
		return nil, err
	}
	out := &smp.OSResetResponse{}
	if err := smp.DecodeMessageBody(resp, out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetMcuMgrParameters retrieves the MCUmgr parameter table, such as buffer size and count, from the device.
func (p *Port) GetMcuMgrParameters(ctx context.Context) (*smp.OSMcuMgrParametersResponse, error) {
	msg, err := smp.NewRequest(smp.OperationRead, smp.GroupOs, smp.OsCommandParameters, &smp.OSMcuMgrParametersRequest{})
	if err != nil {
		return nil, err
	}
	resp, err := p.writeAndRead(ctx, msg)
	if err != nil {
		return nil, err
	}
	out := &smp.OSMcuMgrParametersResponse{}
	if err := smp.DecodeMessageBody(resp, out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetOSAppInfo asks for the device info including kernel version, hardware version, etc.
func (p *Port) GetOSAppInfo(ctx context.Context, format string) (*smp.OSAppInfoResponse, error) {
	msg, err := smp.NewRequest(smp.OperationRead, smp.GroupOs, smp.OsCommandOsAppInfo, &smp.OSAppInfoRequest{
		Format: &format,
	})
	if err != nil {
		return nil, err
	}
	resp, err := p.writeAndRead(ctx, msg)
	if err != nil {
		return nil, err
	}
	out := &smp.OSAppInfoResponse{}
	if err := smp.DecodeMessageBody(resp, out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetBootloaderInfo returns the name of the running bootloader.
func (p *Port) GetBootloaderInfo(ctx context.Context) (*smp.OSBootloaderInfoResponse, error) {
	msg, err := smp.NewRequest(smp.OperationRead, smp.GroupOs, smp.OsCommandBootloaderInfo, &smp.OSBootloaderInfoRequest{})
	if err != nil {
		return nil, err
	}
	resp, err := p.writeAndRead(ctx, msg)
	if err != nil {
		return nil, err
	}
	out := &smp.OSBootloaderInfoResponse{}
	if err := smp.DecodeMessageBody(resp, out); err != nil {
		return nil, err
	}
	return out, nil
}
