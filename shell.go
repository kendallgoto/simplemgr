package simplemgr

import (
	"context"

	"github.com/kendallgoto/simplemgr/pkg/smp"
)

// Execute runs a command as if it were entered into the local shell.
func (p *Port) Execute(ctx context.Context, argv []string) (*smp.ShellExecResponse, error) {
	msg, err := smp.NewRequest(smp.OperationWrite, smp.GroupShell, smp.ShellCommandExecute, &smp.ShellExecRequest{
		Argv: argv,
	})
	if err != nil {
		return nil, err
	}
	resp, err := p.writeAndRead(ctx, msg)
	if err != nil {
		return nil, err
	}
	out := &smp.ShellExecResponse{}
	if err := smp.DecodeMessageBody(resp, out); err != nil {
		return nil, err
	}
	return out, nil
}
