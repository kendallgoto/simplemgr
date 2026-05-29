package simplemgr

import (
	"context"

	"github.com/kendallgoto/simplemgr/pkg/smp"
)

// GetStatGroupData retrieves a named statistics group and provides a map of stats fields.
func (p *Port) GetStatGroupData(ctx context.Context, name string) (*smp.StatGroupDataResponse, error) {
	msg, err := smp.NewRequest(smp.OperationRead, smp.GroupStat, smp.StatCommandGroupData, &smp.StatGroupDataRequest{
		Name: name,
	})
	if err != nil {
		return nil, err
	}
	resp, err := p.writeAndRead(ctx, msg)
	if err != nil {
		return nil, err
	}
	out := &smp.StatGroupDataResponse{}
	if err := smp.DecodeMessageBody(resp, out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetStatListGroups returns a list of all the statistics groups registered on the device.
func (p *Port) GetStatListGroups(ctx context.Context) (*smp.StatListGroupsResponse, error) {
	msg, err := smp.NewRequest(smp.OperationRead, smp.GroupStat, smp.StatCommandListGroups, &smp.StatListGroupsRequest{})
	if err != nil {
		return nil, err
	}
	resp, err := p.writeAndRead(ctx, msg)
	if err != nil {
		return nil, err
	}
	out := &smp.StatListGroupsResponse{}
	if err := smp.DecodeMessageBody(resp, out); err != nil {
		return nil, err
	}
	return out, nil
}
