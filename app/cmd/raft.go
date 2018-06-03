package cmd

import (
	"encoding/json"
	"sort"
	"strconv"
	"time"

	"github.com/genzai-io/sliced/app/api"
	"github.com/genzai-io/sliced/common/raft"
	"github.com/genzai-io/sliced/common/redcon"
)

const RaftTimeout = time.Second * 10

func getClusterRaft() api.RaftService {
	c := api.Cluster
	if c == nil {
		return nil
	} else {
		return c.Raft()
	}
}

func getSliceRaft(slice uint16) api.RaftService {
	return nil
}

func getSchemaSliceRaft(schema uint16, slice uint16) api.RaftService {
	return nil
}

// Sets the correct Raft instance for the duration of the connection
func cmdRAFTSLICE(ctx *Context) Command {
	switch len(ctx.Args) {
	default:
		return ERR("expected 0 or 1 param")

	case 1:
		ctx.Conn.SetKind(api.ConnRaft)
		r := getClusterRaft()
		if r == nil {
			return ERR("raft nil")
		}
		ctx.Conn.SetRaft(r)
		return OK()

	case 2:
		//
		ctx.Conn.SetKind(api.ConnRaft)
		slice, err := strconv.Atoi(string(ctx.Args[1]))
		if err != nil {
			return ERR("invalid slice id: " + string(ctx.Args[1]))
		}

		r := getSliceRaft(uint16(slice))
		if r == nil {
			return ERR("raft nil")
		}

		return OK()
	}
}

func cmdRAFTJOINSLAVE(ctx *Context) Command {
	switch len(ctx.Args) {
	case 2:
		r := getClusterRaft()
		if r == nil {
			return ERR("raft nil")
		}

		//if !r.IsLeader() {
		//	return ERR(fmt.Sprintf("not leader: %s", r.Leader()))
		//}

		err := r.Join(string(ctx.Args[1]), false)
		if err != nil {
			return ERR("join: " + err.Error())
		}
		return OK()

	default:
		return ERR("expected 1 param")
	}
}

func cmdRAFTJOIN(ctx *Context) Command {
	switch len(ctx.Args) {
	case 2:
		r := getClusterRaft()
		if r == nil {
			return ERR("raft nil")
		}

		//if !r.IsLeader() {
		//	return ERR(fmt.Sprintf("not leader: %s", r.Leader()))
		//}

		err := r.Join(string(ctx.Args[1]), true)
		if err != nil {
			return ERR("join: " + err.Error())
		}
		return OK()

	default:
		return ERR("expected 1 param")
	}
}

func cmdRAFTDEMOTE(ctx *Context) Command {
	switch len(ctx.Args) {
	case 2:
		r := getClusterRaft()
		if r == nil {
			return ERR("raft nil")
		}

		//if !r.IsLeader() {
		//	return ERR(fmt.Sprintf("not leader: %s", r.Leader()))
		//}

		err := r.Demote(string(ctx.Args[1]))
		if err != nil {
			return ERR("join: " + err.Error())
		}
		return OK()

	default:
		return ERR("expected 1 param")
	}
}

func cmdRAFTLEAVE(ctx *Context) Command {
	switch len(ctx.Args) {
	case 2:
		r := getClusterRaft()
		if r == nil {
			return ERR("raft nil")
		}

		err := r.Leave(string(ctx.Args[1]))
		if err != nil {
			return ERR("join: " + err.Error())
		}
		return OK()

	default:
		return ERR("expected 1 param")
	}
}

//
//
//
func cmdRAFTVOTE(ctx *Context) Command {
	if len(ctx.Args) != 2 {
		return ERR("expected 1 param")
	}

	r := ctx.Conn.Raft()
	if r == nil {
		return ERR("raft nil")
	}

	b, err := r.Vote(nil, ctx.Args)
	if err != nil {
		return ERR("%s" + err.Error())
	}

	return RAW(b)
}

//
//
//
func cmdRAFTAPPEND(ctx *Context) Command {
	if len(ctx.Args) != 2 {
		return ERR("expected 1 param")
	}

	r := ctx.Conn.Raft()
	if r == nil {
		return ERR("raft nil")
	}

	b, err := r.Append(nil, ctx.Args)
	if err != nil {
		return ERR("%s" + err.Error())
	}

	return RAW(b)
}

//
//
//
func cmdRAFTINSTALL(ctx *Context) Command {
	switch len(ctx.Args) {
	default:
		return ERR("expected 2 or 3 params")
	case 2:
		r := getClusterRaft()
		if r == nil {
			return ERR("raft nil")
		}
		return r.Install(ctx, ctx.Args[1])

	case 3:
		slice, err := strconv.Atoi(string(ctx.Args[1]))
		if err != nil {
			return ERR("invalid slice id: " + string(ctx.Args[1]))
		}

		r := getSliceRaft(uint16(slice))
		if r == nil {
			return ERR("raft nil")
		}

		return r.Install(ctx, ctx.Args[2])
	}
}

//
//
//
func cmdRAFTLEADER(ctx *Context) Command {
	switch len(ctx.Args) {
	default:
		return ERR("expected 0 or 1 param")

	case 1:
		r := getClusterRaft()
		if r == nil {
			return ERR("raft nil")
		}
		return BulkString(string(r.Leader()))

	case 2:
		// Parse slice
		slice, err := strconv.Atoi(string(ctx.Args[1]))
		if err != nil {
			return ERR("invalid slice id: " + string(ctx.Args[1]))
		}

		r := getSliceRaft(uint16(slice))
		if r == nil {
			return ERR("raft nil")
		}

		return BulkString(string(r.Leader()))
	}
}

//
//
//
func cmdRAFTSTATE(ctx *Context) Command {
	switch len(ctx.Args) {
	default:
		return ERR("expected 0 or 1 param")

	case 1:
		r := getClusterRaft()
		if r == nil {
			return ERR("raft nil")
		}
		return BulkString(r.State().String())

	case 2:
		// Parse slice
		slice, err := strconv.Atoi(string(ctx.Args[1]))
		if err != nil {
			return ERR("invalid slice id: " + string(ctx.Args[1]))
		}

		r := getSliceRaft(uint16(slice))
		if r == nil {
			return ERR("raft nil")
		}

		return BulkString(r.State().String())
	}
}

func appendConfig(future raft.ConfigurationFuture) Command {
	var out []byte
	out = redcon.AppendArray(out, len(future.Configuration().Servers)+1)
	out = redcon.AppendInt(out, int64(future.Index()))
	for _, key := range future.Configuration().Servers {
		j, err := json.Marshal(key)
		//out = redcon.AppendBulkString(out, string(key.ID))
		if err != nil {
			out = redcon.AppendBulkString(out, ""+err.Error())
		} else {
			out = redcon.AppendBulk(out, j)
		}
	}

	return RAW(out)
}

//
//
//
func cmdRAFTCONFIG(ctx *Context) Command {
	switch len(ctx.Args) {
	default:
		return ERR("expected 0 or 1 param")

	case 1:
		r := getClusterRaft()
		if r == nil {
			return ERR("raft nil")
		}

		future, err := r.Configuration()
		if err != nil {
			return ERR("" + err.Error())
		}
		return appendConfig(future)

	case 2:
		// Parse slice
		slice, err := strconv.Atoi(string(ctx.Args[1]))
		if err != nil {
			return ERR("invalid slice id: " + string(ctx.Args[1]))
		}

		r := getSliceRaft(uint16(slice))
		if r == nil {
			return ERR("raft nil")
		}

		future, err := r.Configuration()
		if err != nil {
			return ERR("" + err.Error())
		}
		return appendConfig(future)
	}
}

func cmdRAFTSTATS(ctx *Context) Command {
	r := api.Cluster.Raft()

	if r == nil {
		return RAW(ctx.AppendNull())
	}

	stats := r.Stats()
	if stats == nil {
		return RAW(ctx.AppendNull())
	}

	keys := make([]string, 0, len(stats))
	for key := range stats {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var out []byte
	out = redcon.AppendArray(out, len(keys)*2)
	for _, key := range keys {
		j, err := json.Marshal(stats[key])

		out = redcon.AppendBulkString(out, key)
		if err != nil {
			out = redcon.AppendBulkString(out, ""+err.Error())
		} else {
			out = redcon.AppendBulk(out, j)
		}
	}

	return RAW(out)
}
