package simplemgr

import (
	"context"

	goutil "github.com/kendallgoto/goutil"
	"github.com/kendallgoto/simplemgr/pkg/smp"
)

// ReadSetting returns the setting value by name from the device.
func (p *Port) ReadSetting(ctx context.Context, name string, maxSize uint64) (*smp.SettingsReadResponse, error) {
	msg, err := smp.NewRequest(smp.OperationRead, smp.GroupSettings, smp.SettingsCommandReadWrite, &smp.SettingsReadRequest{
		Name:    name,
		MaxSize: goutil.Ptr(maxSize),
	})
	if err != nil {
		return nil, err
	}
	resp, err := p.writeAndRead(ctx, msg)
	if err != nil {
		return nil, err
	}
	out := &smp.SettingsReadResponse{}
	if err := smp.DecodeMessageBody(resp, out); err != nil {
		return nil, err
	}
	return out, nil
}

// WriteSetting sets a new setting value on the device.
func (p *Port) WriteSetting(ctx context.Context, name string, value []byte) (*smp.SettingsWriteResponse, error) {
	msg, err := smp.NewRequest(smp.OperationWrite, smp.GroupSettings, smp.SettingsCommandReadWrite, &smp.SettingsWriteRequest{
		Name:  name,
		Value: value,
	})
	if err != nil {
		return nil, err
	}
	resp, err := p.writeAndRead(ctx, msg)
	if err != nil {
		return nil, err
	}
	out := &smp.SettingsWriteResponse{}
	if err := smp.DecodeMessageBody(resp, out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteSetting deletes an existing setting value from the device, restoring it to the default.
func (p *Port) DeleteSetting(ctx context.Context, name string) (*smp.SettingsDeleteResponse, error) {
	msg, err := smp.NewRequest(smp.OperationWrite, smp.GroupSettings, smp.SettingsCommandDelete, &smp.SettingsDeleteRequest{
		Name: name,
	})
	if err != nil {
		return nil, err
	}
	resp, err := p.writeAndRead(ctx, msg)
	if err != nil {
		return nil, err
	}
	out := &smp.SettingsDeleteResponse{}
	if err := smp.DecodeMessageBody(resp, out); err != nil {
		return nil, err
	}
	return out, nil
}

// CommitSettings makes any changed settings active.
func (p *Port) CommitSettings(ctx context.Context) (*smp.SettingsCommitResponse, error) {
	msg, err := smp.NewRequest(smp.OperationWrite, smp.GroupSettings, smp.SettingsCommandCommit, &smp.SettingsCommitRequest{})
	if err != nil {
		return nil, err
	}
	resp, err := p.writeAndRead(ctx, msg)
	if err != nil {
		return nil, err
	}
	out := &smp.SettingsCommitResponse{}
	if err := smp.DecodeMessageBody(resp, out); err != nil {
		return nil, err
	}
	return out, nil
}

// LoadSettings forces the device to reload settings from flash.
func (p *Port) LoadSettings(ctx context.Context) (*smp.SettingsLoadSaveResponse, error) {
	msg, err := smp.NewRequest(smp.OperationRead, smp.GroupSettings, smp.SettingsCommandLoadSave, &smp.SettingsLoadSaveRequest{})
	if err != nil {
		return nil, err
	}
	resp, err := p.writeAndRead(ctx, msg)
	if err != nil {
		return nil, err
	}
	out := &smp.SettingsLoadSaveResponse{}
	if err := smp.DecodeMessageBody(resp, out); err != nil {
		return nil, err
	}
	return out, nil
}

// SaveSettings persists any modified settings to the flash.
func (p *Port) SaveSettings(ctx context.Context) (*smp.SettingsLoadSaveResponse, error) {
	msg, err := smp.NewRequest(smp.OperationWrite, smp.GroupSettings, smp.SettingsCommandLoadSave, &smp.SettingsLoadSaveRequest{})
	if err != nil {
		return nil, err
	}
	resp, err := p.writeAndRead(ctx, msg)
	if err != nil {
		return nil, err
	}
	out := &smp.SettingsLoadSaveResponse{}
	if err := smp.DecodeMessageBody(resp, out); err != nil {
		return nil, err
	}
	return out, nil
}
