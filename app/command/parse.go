package command

import "github.com/slice-d/genzai/app/api"

// Parse and process next command
func (h *handler) Parse(ctx *Context) Command {
	factory, ok := api.Commands[ctx.Name]
	if !ok {
		return parseOld(ctx)
	}

	return factory.Parse(ctx)
}

func parseOld(ctx *Context) Command {
	switch ctx.Name {
	default:
		return ERR("invalid command")

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
