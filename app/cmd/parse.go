package cmd

import (
	"github.com/genzai-io/sliced/app/api"
	"strings"
)

// Parse and process next command
func (h *handler) Parse(ctx *Context) Command {
	factory, ok := api.Commands[string(ctx.Args[0])]
	if !ok {
		return parseOld(ctx)
	}

	return factory.Parse(ctx.Args)
}

func parseOld(ctx *Context) Command {
	switch strings.ToLower(string(ctx.Args[0])) {
	default:
		return ERR("ERR invalid command")

	case "SLEEP":
		return &Sleep{}

	case api.RaftAppendName:
		return cmdRAFTAPPEND(ctx)

	case api.RaftInstallSnapshotName:
		return cmdRAFTINSTALL(ctx)

	case api.RaftSlice:
		return cmdRAFTSLICE(ctx)

	case api.RaftVoteName:
		return cmdRAFTVOTE(ctx)

	case api.RaftJoinSlaveName:
		return cmdRAFTJOINSLAVE(ctx)

	case api.RaftDemote:
		return cmdRAFTDEMOTE(ctx)

	case api.RaftRemoveName:
		return cmdRAFTLEAVE(ctx)

	case api.RaftStateName:
		return cmdRAFTSTATE(ctx)

	case api.RaftConfigName:
		return cmdRAFTCONFIG(ctx)

	case api.RaftStatsName:
		return cmdRAFTSTATS(ctx)
	}
}
