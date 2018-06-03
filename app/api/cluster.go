package api

var Cluster ICluster

type ICluster interface {
	Raft() RaftService
}

