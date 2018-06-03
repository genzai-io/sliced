// Provides a raft transport using the Redis RESP protocol.
package core

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/genzai-io/sliced"
	"github.com/genzai-io/sliced/app/api"
	cmd "github.com/genzai-io/sliced/app/cmd"
	"github.com/genzai-io/sliced/common/raft"
	"github.com/genzai-io/sliced/common/redcon"
	"github.com/genzai-io/sliced/common/service"
)

const (
	// DefaultTimeoutScale is the default TimeoutScale in a NetworkTransport.
	DefaultTimeoutScale = 256 * 1024 // 256KB

	// rpcMaxPipeline controls the maximum number of outstanding
	// Append RPC calls.
	rpcMaxPipeline = 128
)

var (
	errInvalidNumberOfArgs = errors.New("invalid number or arguments")
	errInvalidCommand      = errors.New("invalid cmd")
	errInvalidRequest      = errors.New("invalid request")
	errInvalidResponse     = errors.New("invalid response")
)

// Raft Transport that uses Redis protocol over main API event loops.
type RaftTransport struct {
	service.BaseService

	schemaID int32
	sliceID  int32
	addr     raft.ServerAddress
	consumer chan raft.RPC

	raftInstallBytes []byte

	mu     sync.Mutex
	pools  map[string]*redis.Pool
	closed bool
}

func NewTransport(schemaID int32, sliceID int32, loggerName string) *RaftTransport {
	t := &RaftTransport{
		schemaID: schemaID,
		sliceID:  sliceID,
		addr:     moved.ClusterAddress,
		consumer: make(chan raft.RPC),
		pools:    make(map[string]*redis.Pool),
	}

	t.BaseService = *service.NewBaseService(moved.Logger, loggerName, t)
	return t
}

func (t *RaftTransport) OnStart() error {
	return nil
}

func (t *RaftTransport) OnStop() {
	if err := t.Close(); err != nil {
		t.Logger.Error().Err(err)
	}
}

// newTargetPool returns a Redigo pool for the specified target node.
func (t *RaftTransport) newTargetPool(target string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     5,           // figure 5 should suffice most clusters.
		IdleTimeout: time.Minute, //
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", target)
			if err != nil {
				return nil, err
			}
			// Build the connection for the correct slice
			if t.schemaID < 0 {
				reply, err := c.Do("RSLICE")
				if err != nil {
					return nil, err
				}
				_ = reply
				return c, nil
			} else {
				reply, err := c.Do("RSLICE", t.schemaID, t.sliceID)
				if err != nil {
					return nil, err
				}
				_ = reply
				return c, nil
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}

// Close is used to permanently disable the transport
func (t *RaftTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return errors.New("closed")
	}
	t.closed = true
	for _, pool := range t.pools {
		pool.Close()
	}
	t.pools = nil
	return nil
}

// getPool returns a usable pool for obtaining a connection to the specified target.
func (t *RaftTransport) getPool(target string) (*redis.Pool, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return nil, errors.New("closed")
	}
	pool, ok := t.pools[target]
	if !ok {
		pool = t.newTargetPool(target)
		t.pools[target] = pool
	}
	return pool, nil
}

// getConn returns a connection to the target.
func (t *RaftTransport) getConn(target string) (redis.Conn, error) {
	pool, err := t.getPool(target)
	if err != nil {
		return nil, err
	}
	return pool.Get(), nil
}

// AppendEntriesPipeline returns an interface that can be used to pipeline Append requests.
func (t *RaftTransport) AppendEntriesPipeline(id raft.ServerID, target raft.ServerAddress) (raft.AppendPipeline, error) {
	//// Get a connection
	//conn, err := t.getConn(string(target))
	//if err != nil {
	//	return nil, raft.ErrPipelineReplicationNotSupported
	//}
	//
	//// Create the pipeline
	//return newNetPipeline(t, conn), nil
	return nil, raft.ErrPipelineReplicationNotSupported
}

// encodeAppendEntriesRequest encodes AppendEntriesRequest arguments into a
// tight binary format.
func encodeAppendEntriesRequest(args *raft.AppendEntriesRequest) []byte {
	n := make([]byte, 8)       // used to store uint64s
	b := make([]byte, 42, 256) // encoded message goes here

	binary.LittleEndian.PutUint16(b[0:2], uint16(args.ProtocolVersion))
	binary.LittleEndian.PutUint64(b[2:10], args.Term)
	binary.LittleEndian.PutUint64(b[10:18], args.PrevLogEntry)
	binary.LittleEndian.PutUint64(b[18:26], args.PrevLogTerm)
	binary.LittleEndian.PutUint64(b[26:34], args.LeaderCommitIndex)
	binary.LittleEndian.PutUint64(b[34:42], uint64(len(args.Leader)))
	b = append(b, args.Leader...)
	binary.LittleEndian.PutUint64(n, uint64(len(args.Entries)))
	b = append(b, n...)
	for _, entry := range args.Entries {
		binary.LittleEndian.PutUint64(n, entry.Index)
		b = append(b, n...)
		binary.LittleEndian.PutUint64(n, entry.Term)
		b = append(b, n...)
		b = append(b, byte(entry.Type))
		binary.LittleEndian.PutUint64(n, uint64(len(entry.Data)))
		b = append(b, n...)
		b = append(b, entry.Data...)
	}
	return b
}

// decodeAppendEntriesRequest decodes AppendEntriesRequest data.
// Returns true when successful
func decodeAppendEntriesRequest(b []byte, args *raft.AppendEntriesRequest) bool {
	if len(b) < 42 {
		return false
	}
	args.ProtocolVersion = raft.ProtocolVersion(binary.LittleEndian.Uint16(b[0:2]))
	args.Term = binary.LittleEndian.Uint64(b[2:10])
	args.PrevLogEntry = binary.LittleEndian.Uint64(b[10:18])
	args.PrevLogTerm = binary.LittleEndian.Uint64(b[18:26])
	args.LeaderCommitIndex = binary.LittleEndian.Uint64(b[26:34])
	args.Leader = make([]byte, int(binary.LittleEndian.Uint64(b[34:42])))
	b = b[42:]
	if len(b) < len(args.Leader) {
		return false
	}
	copy(args.Leader, b[:len(args.Leader)])
	b = b[len(args.Leader):]
	if len(b) < 8 {
		return false
	}
	args.Entries = make([]*raft.Log, int(binary.LittleEndian.Uint64(b)))
	b = b[8:]
	for i := 0; i < len(args.Entries); i++ {
		if len(b) < 25 {
			return false
		}
		args.Entries[i] = &raft.Log{}
		args.Entries[i].Index = binary.LittleEndian.Uint64(b[0:8])
		args.Entries[i].Term = binary.LittleEndian.Uint64(b[8:16])
		args.Entries[i].Type = raft.LogType(b[16])
		args.Entries[i].Data = make([]byte, int(binary.LittleEndian.Uint64(b[17:25])))
		b = b[25:]
		if len(b) < len(args.Entries[i].Data) {
			return false
		}
		copy(args.Entries[i].Data, b[:len(args.Entries[i].Data)])
		b = b[len(args.Entries[i].Data):]
	}
	return len(b) == 0
}

func encodeAppendEntriesResponse(args *raft.AppendEntriesResponse) []byte {
	b := make([]byte, 20)
	binary.LittleEndian.PutUint16(b[0:2], uint16(args.ProtocolVersion))
	binary.LittleEndian.PutUint64(b[2:10], args.Term)
	binary.LittleEndian.PutUint64(b[10:18], args.LastLog)
	if args.Success {
		b[18] = 1
	}
	if args.NoRetryBackoff {
		b[19] = 1
	}
	return b
}

func decodeAppendEntriesResponse(b []byte, args *raft.AppendEntriesResponse) bool {
	if len(b) != 20 {
		return false
	}
	args.ProtocolVersion = raft.ProtocolVersion(binary.LittleEndian.Uint16(b[0:2]))
	args.Term = binary.LittleEndian.Uint64(b[2:10])
	args.LastLog = binary.LittleEndian.Uint64(b[10:18])
	if b[18] == 1 {
		args.Success = true
	} else {
		args.Success = false
	}
	if b[19] == 1 {
		args.NoRetryBackoff = true
	} else {
		args.NoRetryBackoff = false
	}
	return true
}

// Append implements the Transport interface.
func (t *RaftTransport) AppendEntries(id raft.ServerID, target raft.ServerAddress, args *raft.AppendEntriesRequest, resp *raft.AppendEntriesResponse) error {
	conn, err := t.getConn(string(target))
	if err != nil {
		return err
	}
	defer conn.Close()

	reply, err := conn.Do(api.RaftAppendName, encodeAppendEntriesRequest(args))
	if err != nil {
		return err
	}
	switch val := reply.(type) {
	default:
		return errInvalidResponse
	case redis.Error:
		return val
	case []byte:
		if !decodeAppendEntriesResponse(val, resp) {
			return errInvalidResponse
		}
		return nil
	}
}

func (t *RaftTransport) handleAppendEntries(o []byte, args [][]byte) ([]byte, error) {
	if len(args) != 2 {
		return redcon.AppendError(o, ""+errInvalidNumberOfArgs.Error()), errInvalidNumberOfArgs
	}
	var rpc raft.RPC
	var aer raft.AppendEntriesRequest
	if !decodeAppendEntriesRequest(args[1], &aer) {
		return redcon.AppendError(o, "invalid request"), errInvalidRequest
	}
	rpc.Command = &aer
	respChan := make(chan raft.RPCResponse)
	rpc.RespChan = respChan
	t.consumer <- rpc
	rresp := <-respChan
	if rresp.Error != nil {
		return redcon.AppendError(o, ""+rresp.Error.Error()), rresp.Error
	}
	resp, ok := rresp.Response.(*raft.AppendEntriesResponse)
	if !ok {
		return redcon.AppendError(o, "invalid response"), errInvalidResponse
	}
	data := encodeAppendEntriesResponse(resp)
	return redcon.AppendBulk(o, data), nil
}

// Vote implements the Transport interface.
func (t *RaftTransport) RequestVote(id raft.ServerID, target raft.ServerAddress, args *raft.RequestVoteRequest, resp *raft.RequestVoteResponse) error {
	conn, err := t.getConn(string(target))
	if err != nil {
		return err
	}
	defer conn.Close()

	data, _ := json.Marshal(args)
	//reply, _, err := Do(string(target), nil, api.RaftVote, data)
	reply, err := conn.Do(api.RaftVoteName, data)
	//val, _, err := Do(string(target), nil, []byte("RAFTVOTE"), data)
	if err != nil {
		t.Logger.Error().AnErr("err", err).Msgf("VOTE to %s error", string(target))
		return err
	}

	switch val := reply.(type) {
	default:
		return errors.New("invalid response")
	case redis.Error:
		return val
	case []byte:
		if err := json.Unmarshal(val, resp); err != nil {
			return err
		}
		return nil
	}

	return nil
}

func (t *RaftTransport) handleRequestVote(o []byte, args [][]byte) ([]byte, error) {
	if len(args) != 2 {
		return redcon.AppendError(o, ""+errInvalidNumberOfArgs.Error()), errInvalidNumberOfArgs
	}
	var aer raft.RequestVoteRequest
	if err := json.Unmarshal(args[1], &aer); err != nil {
		return redcon.AppendError(o, ""+err.Error()), err
	}

	respCh := make(chan raft.RPCResponse, 1)
	rpc := raft.RPC{
		RespChan: respCh,
	}
	rpc.Command = &aer

	t.consumer <- rpc
	rresp := <-respCh
	if rresp.Error != nil {
		return redcon.AppendError(o, ""+rresp.Error.Error()), rresp.Error
	}
	resp, ok := rresp.Response.(*raft.RequestVoteResponse)
	if !ok {
		return redcon.AppendError(o, "invalid response"), errors.New("invalid response")
	}
	data, _ := json.Marshal(resp)
	return redcon.AppendBulk(o, data), nil
}

// InstallSnapshot implements the Transport interface.
func (t *RaftTransport) InstallSnapshot(
	id raft.ServerID, target raft.ServerAddress, args *raft.InstallSnapshotRequest, resp *raft.InstallSnapshotResponse, data io.Reader,
) error {
	// Use a dedicated connection for snapshots. This operation happens very infrequently, but when it does
	// it often passes a lot of data.
	conn, err := net.Dial("tcp", string(target))
	if err != nil {
		return err
	}
	defer conn.Close()

	rd := bufio.NewReader(conn)
	// use JSON encoded arguments for the initial request.
	rdata, err := json.Marshal(args)
	if err != nil {
		return err
	}
	if t.sliceID < 0 {
		// send RAFTINSTALL {args}
		if _, err := conn.Write(buildCommand(nil, api.RaftInstall, rdata)); err != nil {
			return err
		}
	} else {
		// send RAFTINSTALL {slice} {args}
		if _, err := conn.Write(buildCommand(nil, api.RaftInstall, []byte(fmt.Sprintf("%d", t.sliceID)), rdata)); err != nil {
			return err
		}
	}

	// receive +OK
	line, err := response(rd)
	if err != nil {
		return err
	}
	if string(line) != "OK" {
		return errInvalidResponse
	}
	var i int
	var cmd []byte                   // reuse buffer
	buf := make([]byte, 4*1024*1024) // 4MB chunk
	for {
		n, ferr := data.Read(buf)
		if n > 0 {
			// send CHUNK data
			cmd = buildCommand(cmd, api.RaftChunk, buf[:n])
			if _, err := conn.Write(cmd); err != nil {
				return err
			}
			cmd = cmd[:0] // set len to zero for reuse
			// receive +OK
			line, err := response(rd)
			if err != nil {
				return err
			}
			if string(line) != "OK" {
				return errInvalidResponse
			}
			i++
		}
		if ferr != nil {
			if ferr == io.EOF {
				break
			}
			return ferr
		}
	}
	// send DONE
	if _, err := conn.Write(buildCommand(nil, api.RaftDone)); err != nil {
		return err
	}
	// receive ${resp}
	line, err = response(rd)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(line, resp); err != nil {
		return err
	}
	return nil
}

type snapshotHandler struct {
	transport  *RaftTransport
	time       time.Time
	reader     *io.PipeReader
	writer     *io.PipeWriter
	handler    api.IHandler
	downstream uint64
}

func (s *snapshotHandler) Commit(ctx *cmd.Context) {
}

func (s *snapshotHandler) Parse(ctx *cmd.Context) cmd.Command {
	args := ctx.Args
	conn := ctx.Conn
	packet := ctx.Packet

	switch ctx.Name {
	default:
		s.reader.CloseWithError(errInvalidRequest)
		s.writer.CloseWithError(errInvalidRequest)
		conn.Close()
		return cmd.ERR(fmt.Sprintf("install snapshot mode only supports '%s' and '%s'", api.RaftChunkName, api.RaftDoneName))
	case api.RaftChunkName:
		s.downstream += uint64(len(packet))

		if len(args) != 2 {
			s.reader.CloseWithError(errInvalidNumberOfArgs)
			s.writer.CloseWithError(errInvalidNumberOfArgs)
			conn.Close()
			return cmd.ERR("invalid number of args")
		}
		if _, err := s.writer.Write(args[1]); err != nil {
			s.reader.CloseWithError(err)
			s.writer.CloseWithError(err)
			conn.Close()
			return cmd.ERR(fmt.Sprintf("writer error '%s'", err.Error()))
		}
		return cmd.OK()
	case api.RaftDoneName:
		s.writer.Close()
		s.reader.Close()
		return cmd.OK()
	}
	return nil
}

func (t *RaftTransport) HandleInstallSnapshot(ctx *cmd.Context, arg []byte) cmd.Command {
	var rpc raft.RPC
	rpc.Command = &raft.InstallSnapshotRequest{}
	if err := json.Unmarshal(arg, &rpc.Command); err != nil {
		return cmd.ERROR(err)
	}

	// Create new pipe
	rd, wr := io.Pipe()

	// SortedSet rpc reader
	rpc.Reader = rd
	respChan := make(chan raft.RPCResponse)
	rpc.RespChan = respChan

	// Send to transport consumer
	t.consumer <- rpc

	// Wait for response
	resp := <-respChan
	if resp.Error != nil {
		rd.Close()
		wr.Close()
		return cmd.ERROR(resp.Error)
	}
	// Marshal response
	data, err := json.Marshal(resp.Response)
	if err != nil {
		rd.Close()
		wr.Close()
		return cmd.ERROR(err)
	}

	handler := &snapshotHandler{
		transport: t,
		time:      time.Now(),
		reader:    rd,
		writer:    wr,
	}
	handler.handler = ctx.Conn.SetHandler(handler)

	out := redcon.AppendOK(nil)
	out = redcon.AppendBulk(out, data)
	return cmd.RAW(out)
}

//func (t *RESPTransport) handleInstallSnapshot(conn redcon.DetachedConn, arg []byte) {
//	err := func() error {
//		var rpc raft.RPC
//		rpc.Command = &raft.InstallSnapshotRequest{}
//		if err := json.Unmarshal(arg, &rpc.Command); err != nil {
//			return err
//		}
//		conn.WriteString("OK")
//		if err := conn.Flush(); err != nil {
//			return err
//		}
//		rd, wr := io.Pipe()
//		go func() {
//			err := func() error {
//				var i int
//				for {
//					cmd, err := conn.ReadCommand()
//					if err != nil {
//						return err
//					}
//					switch strings.ToUpper(string(cmd.Args[0])) {
//					default:
//						return errInvalidCommand
//					case RaftChunkName:
//						if len(cmd.Args) != 2 {
//							return errInvalidNumberOfArgs
//						}
//						if _, err := wr.Write(cmd.Args[1]); err != nil {
//							return err
//						}
//						conn.WriteString("OK")
//						if err := conn.Flush(); err != nil {
//							return err
//						}
//						i++
//					case RaftDoneName:
//						return nil
//					}
//				}
//			}()
//			if err != nil {
//				wr.CloseWithError(err)
//			} else {
//				wr.Close()
//			}
//		}()
//		rpc.Reader = rd
//		respChan := make(chan raft.RPCResponse)
//		rpc.RespChan = respChan
//		t.consumer <- rpc
//		resp := <-respChan
//		if resp.Error != nil {
//			return resp.Error
//		}
//		data, err := json.Marshal(resp.Response)
//		if err != nil {
//			return err
//		}
//		conn.WriteBulk(data)
//		if err := conn.Flush(); err != nil {
//			return err
//		}
//		return nil
//	}()
//	if err != nil {
//		t.Logger.Warn().Msgf("handle install snapshot failed: %v", err)
//	} else {
//		t.Logger.Info().Msg("handle install snapshot completed")
//	}
//}

//func (t *RESPTransport) handle(conn redcon.Conn, cmd redcon.Command) {
//	var err error
//	var res []byte
//	switch strings.ToLower(string(cmd.Args[0])) {
//	default:
//		if t.handleFn != nil {
//			t.handleFn(conn, cmd)
//		} else {
//			conn.WriteError("unknown Cmd '" + string(cmd.Args[0]) + "'")
//		}
//		return
//	case "raftinstallsnapshot":
//		if len(cmd.Args) != 2 {
//			err = errInvalidNumberOfArgs
//		} else {
//			// detach connection and forward to the background
//			dconn := conn.Detach()
//			go func() {
//				defer dconn.Close()
//				t.Install(dconn, cmd.Args[1])
//			}()
//			return
//		}
//	case "raftrequestvote":
//		res, err = t.handleRequestVote(cmd)
//	case "raftappendentries":
//		res, err = t.handleAppendEntries(cmd)
//	}
//	if err != nil {
//		if err == errInvalidNumberOfArgs {
//			conn.WriteError("wrong number of arguments for '" + string(cmd.Args[0]) + "' Cmd")
//		} else {
//			conn.WriteError("" + err.Error())
//		}
//	} else {
//		conn.WriteBulk(res)
//	}
//}

// Consumer implements the Transport interface.
func (t *RaftTransport) Consumer() <-chan raft.RPC { return t.consumer }

// LocalAddr implements the Transport interface.
func (t *RaftTransport) LocalAddr() raft.ServerAddress { return t.addr }

// EncodePeer implements the Transport interface.
func (t *RaftTransport) EncodePeer(id raft.ServerID, peer raft.ServerAddress) []byte { return []byte(peer) }

// DecodePeer implements the Transport interface.
func (t *RaftTransport) DecodePeer(peer []byte) raft.ServerAddress { return raft.ServerAddress(peer) }

// SetHeartbeatHandler implements the Transport interface.
func (t *RaftTransport) SetHeartbeatHandler(cb func(rpc raft.RPC)) {}

// Do is a helper function that makes a very simple remote request with
// the specified Cmd.
// The addr param is the target server address.
// The buf param is an optional reusable buffer, this can be nil.
// The args are the Cmd arguments such as "SET", "key", "value".
// Return response is a bulk, string, or an error.
// The nbuf is a reuseable buffer, this can be ignored.
func Do(addr string, buf []byte, args ...[]byte) (resp []byte, nbuf []byte, err error) {
	cmd := buildCommand(buf, args...)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, cmd, err
	}
	defer conn.Close()
	if _, err = conn.Write(cmd); err != nil {
		return nil, cmd, err
	}
	resp, err = response(bufio.NewReader(conn))
	return resp, cmd[:0], err
}

func response(rd *bufio.Reader) ([]byte, error) {
	c, err := rd.ReadByte()
	if err != nil {
		return nil, err
	}
	switch c {
	default:
		return nil, errors.New("invalid response")
	case '+', '-', '$', ':', '*':
		line, err := rd.ReadBytes('\n')
		if err != nil {
			return nil, err
		}
		if len(line) < 2 || line[len(line)-2] != '\r' {
			return nil, errors.New("invalid response")
		}
		line = line[:len(line)-2]
		switch c {
		default:
			return nil, errors.New("invalid response")
		case '*':
			n, err := strconv.ParseUint(string(line), 10, 64)
			if err != nil {
				return nil, err
			}
			var buf []byte
			for i := 0; i < int(n); i++ {
				res, err := response(rd)
				if err != nil {
					return nil, err
				}
				buf = append(buf, res...)
				buf = append(buf, "\n"...)
			}
			return buf, nil
		case '+', ':':
			return line, nil
		case '-':
			return nil, errors.New(string(line))
		case '$':
			n, err := strconv.ParseUint(string(line), 10, 64)
			if err != nil {
				return nil, err
			}
			data := make([]byte, int(n)+2)
			if _, err := io.ReadFull(rd, data); err != nil {
				return nil, err
			}
			if data[len(data)-2] != '\r' || data[len(data)-1] != '\n' {
				return nil, errors.New("invalid response")
			}
			return data[:len(data)-2], nil
		}
	}
}

// buildCommand builds a valid redis Cmd and appends to buf.
// The return value is the newly appended buf.
func buildCommand(buf []byte, args ...[]byte) []byte {
	buf = append(buf, '*')
	buf = append(buf, strconv.FormatInt(int64(len(args)), 10)...)
	buf = append(buf, '\r', '\n')
	for _, arg := range args {
		buf = append(buf, '$')
		buf = append(buf, strconv.FormatInt(int64(len(arg)), 10)...)
		buf = append(buf, '\r', '\n')
		buf = append(buf, arg...)
		buf = append(buf, '\r', '\n')
	}
	return buf
}

func ReadRawResponse(rd *bufio.Reader) (raw []byte, kind byte, err error) {
	kind, err = rd.ReadByte()
	if err != nil {
		return raw, kind, err
	}
	raw = append(raw, kind)
	switch kind {
	default:
		return raw, kind, errors.New("invalid response")
	case '+', '-', '$', ':', '*':
		line, err := rd.ReadBytes('\n')
		if err != nil {
			return raw, kind, err
		}
		raw = append(raw, line...)
		if len(line) < 2 || line[len(line)-2] != '\r' {
			return raw, kind, errors.New("invalid response")
		}
		line = line[:len(line)-2]
		switch kind {
		default:
			return raw, kind, errors.New("invalid response")
		case '+', ':', '-':
			return raw, kind, nil
		case '*':
			n, err := strconv.ParseInt(string(line), 10, 64)
			if err != nil {
				return raw, kind, err
			}
			if n > 0 {
				for i := 0; i < int(n); i++ {
					res, _, err := ReadRawResponse(rd)
					if err != nil {
						return raw, kind, err
					}
					raw = append(raw, res...)
				}
			}
		case '$':
			n, err := strconv.ParseInt(string(line), 10, 64)
			if err != nil {
				return raw, kind, err
			}
			if n > 0 {
				data := make([]byte, int(n)+2)
				if _, err := io.ReadFull(rd, data); err != nil {
					return raw, kind, err
				}
				if data[len(data)-2] != '\r' || data[len(data)-1] != '\n' {
					return raw, kind, errors.New("invalid response")
				}
				raw = append(raw, data...)
			}
		}
		return raw, kind, nil
	}
}

// deferError can be embedded to allow a future
// to provide an error in the future.
type deferError struct {
	err       error
	errCh     chan error
	responded bool
}

func (d *deferError) init() {
	d.errCh = make(chan error, 1)
}

func (d *deferError) Error() error {
	if d.err != nil {
		// Note that when we've received a nil error, this
		// won't trigger, but the channel is closed after
		// send so we'll still return nil below.
		return d.err
	}
	if d.errCh == nil {
		panic("waiting for response on nil channel")
	}
	d.err = <-d.errCh
	return d.err
}

func (d *deferError) respond(err error) {
	if d.errCh == nil {
		return
	}
	if d.responded {
		return
	}
	d.errCh <- err
	close(d.errCh)
	d.responded = true
}

// appendFuture is used for waiting on a pipelined append
// entries RPC.
type appendFuture struct {
	deferError
	start time.Time
	args  *raft.AppendEntriesRequest
	resp  *raft.AppendEntriesResponse
}

func (a *appendFuture) Start() time.Time {
	return a.start
}

func (a *appendFuture) Request() *raft.AppendEntriesRequest {
	return a.args
}

func (a *appendFuture) Response() *raft.AppendEntriesResponse {
	return a.resp
}

type netPipeline struct {
	schemaID int32
	sliceID  int32
	conn     redis.Conn
	trans    *RaftTransport

	doneCh       chan raft.AppendFuture
	inprogressCh chan *appendFuture

	shutdown     bool
	shutdownCh   chan struct{}
	shutdownLock sync.Mutex
}

// newNetPipeline is used to construct a netPipeline from a given
// transport and connection.
func newNetPipeline(trans *RaftTransport, conn redis.Conn) *netPipeline {
	n := &netPipeline{
		schemaID:     trans.schemaID,
		sliceID:      trans.sliceID,
		conn:         conn,
		trans:        trans,
		doneCh:       make(chan raft.AppendFuture, rpcMaxPipeline),
		inprogressCh: make(chan *appendFuture, rpcMaxPipeline),
		shutdownCh:   make(chan struct{}),
	}
	go n.decodeResponses()
	return n
}

// decodeResponse is used to decode an RPC response and reports whether
// the connection can be reused.
//func decodeResponse(conn redis.Conn, resp interface{}) (bool, error) {
//	// Decode the error if any
//	var rpcError string
//	if err := conn.dec.Decode(&rpcError); err != nil {
//		conn.Close()
//		return false, err
//	}
//
//	// Decode the response
//	if err := conn.dec.Decode(resp); err != nil {
//		conn.Release()
//		return false, err
//	}
//
//	// Fmt an error if any
//	if rpcError != "" {
//		return true, fmt.Errorf(rpcError)
//	}
//	return true, nil
//}

// decodeResponses is a long running routine that decodes the responses
// sent on the connection.
func (n *netPipeline) decodeResponses() {
	//timeout := n.trans.timeout
	for {
		select {
		case future := <-n.inprogressCh:
			//if timeout > 0 {
			//n.conn.SetReadDeadline(time.Now().Add(timeout))
			//}
			reply, err := n.conn.Receive()

			switch val := reply.(type) {
			default:
				future.respond(errInvalidResponse)
			case redis.Error:
				future.respond(err)
			case []byte:
				if !decodeAppendEntriesResponse(val, future.resp) {
					future.respond(err)
				} else {
					future.respond(nil)
				}
			}

			select {
			case n.doneCh <- future:
			case <-n.shutdownCh:
				return
			}
		case <-n.shutdownCh:
			return
		}
	}
}

// Append is used to pipeline a new append entries request.
func (n *netPipeline) AppendEntries(args *raft.AppendEntriesRequest, resp *raft.AppendEntriesResponse) (raft.AppendFuture, error) {
	// Create a new future
	future := &appendFuture{
		start: time.Now(),
		args:  args,
		resp:  resp,
	}
	future.init()

	// Add a send timeout
	//if timeout := n.trans.timeout; timeout > 0 {
	//	n.conn.conn.SetWriteDeadline(time.Now().Add(timeout))
	//}

	err := n.conn.Send(api.RaftAppendName, encodeAppendEntriesRequest(args))
	if err != nil {
		return nil, err
	}

	//// Send the RPC
	//if err := sendRPC(n.conn, rpcAppendEntries, future.args); err != nil {
	//	return nil, err
	//}

	// Hand-off for decoding, this can also cause back-pressure
	// to prevent too many inflight requests
	select {
	case n.inprogressCh <- future:
		return future, nil
	case <-n.shutdownCh:
		return nil, raft.ErrPipelineShutdown
	}
}

// Consumer returns a channel that can be used to consume complete futures.
func (n *netPipeline) Consumer() <-chan raft.AppendFuture {
	return n.doneCh
}

// Closed is used to shutdown the pipeline connection.
func (n *netPipeline) Close() error {
	n.shutdownLock.Lock()
	defer n.shutdownLock.Unlock()
	if n.shutdown {
		return nil
	}

	// Release the connection
	n.conn.Close()

	n.shutdown = true
	close(n.shutdownCh)
	return nil
}
