package rpc

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var (
	RpcSyncSeq     int64
	RpcSyncSeqMap 	sync.Map
	CLUSTER	ICluster
)

type(
	RpcSync struct{
		RpcChan chan RetInfo
		Seq int64
	}

	RetInfo struct {
		Err error
		Ret []interface{}
	}

	RetInfoEx struct {
		Err string
		Ret []interface{}
	}

	ICluster interface {
		SendMsg(RpcHead, string, ...interface{})
	}
)

const(
	MAX_RPC_TIMEOUT = 3*time.Second
)

func (this *RetInfo) ToJson() RetInfoEx{
	ret := RetInfoEx{Err:"", Ret:this.Ret}
	if this.Err != nil{
		ret.Err = this.Err.Error()
	}
	return ret
}

func (this *RetInfoEx) ToJson() RetInfo{
	ret := RetInfo{Err:nil, Ret:this.Ret}
	if this.Err != ""{
		ret.Err = errors.New(this.Err)
	}
	return ret
}

func CrateRpcSync() *RpcSync{
	req := RpcSync{}
	req.Seq = atomic.AddInt64(&RpcSyncSeq, 1)
	req.RpcChan = make(chan RetInfo)
	RpcSyncSeqMap.Store(req.Seq, &req)
	return &req
}

func GetRpcSync(seq int64) *RpcSync{
	req, bOk := RpcSyncSeqMap.Load(seq)
	if bOk{
		RpcSyncSeqMap.Delete(seq)
		return req.(*RpcSync)
	}
	return nil
}

func Sync(seq int64, ret RetInfo) bool{
	req := GetRpcSync(seq)
	if req != nil{
		req.RpcChan <- ret
		return true
	}
	return false
}

//同步Call
/*
	resut := pPlayer.SyncMsg(rpc.RpcHead{}, "rpctest", gateClusterId, zoneClusterId)

	this.RegisterCall("rpctest", func(ctx context.Context, gateClusterId uint32, zoneClusterId uint32) rpc.RetInfo{
		return rpc.RetInfo{nil, []interface{}{gateClusterId, zoneClusterId}}
	})

	this.RegisterCall("test", func(ctx context.Context, a , b int) rpc.RetInfo{
		fmt.Println(a, b)
		return rpc.RetInfo{Err:errors.New("test"), Ret:[]interface{}{a, b}}
	})

	fmt.Println(world.SERVER.GetClusterMgr().SyncMsg(rpc.RpcHead{DestServerType:message.SERVICE_ACCOUNTSERVER, SendType:message.SEND_BALANCE}, "test", 1, 2))
*/