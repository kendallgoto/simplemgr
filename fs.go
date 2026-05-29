package simplemgr

import (
	"context"

	"github.com/kendallgoto/simplemgr/pkg/smp"
)

// DownloadFile asks for a specific file by name and reads the data back, with an offset to seek in the file.
func (p *Port) DownloadFile(ctx context.Context, name string, offset uint64) (*smp.FsDownloadResponse, error) {
	msg, err := smp.NewRequest(smp.OperationRead, smp.GroupFs, smp.FsCommandDownloadUpload, &smp.FsDownloadRequest{
		Name:   name,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}
	resp, err := p.writeAndRead(ctx, msg)
	if err != nil {
		return nil, err
	}
	out := &smp.FsDownloadResponse{}
	if err := smp.DecodeMessageBody(resp, out); err != nil {
		return nil, err
	}
	return out, nil
}

// UploadFile uploads a portion of a file to the remote device.
func (p *Port) UploadFile(ctx context.Context, req *smp.FsUploadRequest) (*smp.FsUploadResponse, error) {
	msg, err := smp.NewRequest(smp.OperationWrite, smp.GroupFs, smp.FsCommandDownloadUpload, req)
	if err != nil {
		return nil, err
	}
	resp, err := p.writeAndRead(ctx, msg)
	if err != nil {
		return nil, err
	}
	out := &smp.FsUploadResponse{}
	if err := smp.DecodeMessageBody(resp, out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetFileStatus returns the file length or an error if it is not present.
func (p *Port) GetFileStatus(ctx context.Context, name string) (*smp.FsStatusResponse, error) {
	msg, err := smp.NewRequest(smp.OperationRead, smp.GroupFs, smp.FsCommandStatus, &smp.FsStatusRequest{
		Name: name,
	})
	if err != nil {
		return nil, err
	}
	resp, err := p.writeAndRead(ctx, msg)
	if err != nil {
		return nil, err
	}
	out := &smp.FsStatusResponse{}
	if err := smp.DecodeMessageBody(resp, out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetFileHash retrieves a hash of the existing file from the device. A specific hashing algorithm can be provided.
func (p *Port) GetFileHash(ctx context.Context, req *smp.FsHashRequest) (*smp.FsHashResponse, error) {
	msg, err := smp.NewRequest(smp.OperationRead, smp.GroupFs, smp.FsCommandHash, req)
	if err != nil {
		return nil, err
	}
	resp, err := p.writeAndRead(ctx, msg)
	if err != nil {
		return nil, err
	}
	out := &smp.FsHashResponse{}
	if err := smp.DecodeMessageBody(resp, out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetSupportedFileHashTypes provides a list of the device's supported hashing algorithms to be used with GetFileHash.
func (p *Port) GetSupportedFileHashTypes(ctx context.Context) (*smp.FsSupportedHashTypesResponse, error) {
	msg, err := smp.NewRequest(smp.OperationRead, smp.GroupFs, smp.FsCommandSupportedHashTypes, &smp.FsSupportedHashTypesRequest{})
	if err != nil {
		return nil, err
	}
	resp, err := p.writeAndRead(ctx, msg)
	if err != nil {
		return nil, err
	}
	out := &smp.FsSupportedHashTypesResponse{}
	if err := smp.DecodeMessageBody(resp, out); err != nil {
		return nil, err
	}
	return out, nil
}

// CloseFile tells the device to close and flush an open file after writing.
func (p *Port) CloseFile(ctx context.Context) (*smp.FsCloseResponse, error) {
	msg, err := smp.NewRequest(smp.OperationWrite, smp.GroupFs, smp.FsCommandClose, &smp.FsCloseRequest{})
	if err != nil {
		return nil, err
	}
	resp, err := p.writeAndRead(ctx, msg)
	if err != nil {
		return nil, err
	}
	out := &smp.FsCloseResponse{}
	if err := smp.DecodeMessageBody(resp, out); err != nil {
		return nil, err
	}
	return out, nil
}
