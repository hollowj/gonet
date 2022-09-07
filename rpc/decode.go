package rpc

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"github.com/golang/protobuf/proto"
	"gonet/base"
	"reflect"
)

//rpc UnmarshalHead
func UnmarshalHead(buff []byte) (*RpcPacket, RpcHead) {
	nLen := base.Clamp(len(buff), 0, 256)
	return Unmarshal(buff[:nLen])
}

func Unmarshal(buff []byte) (*RpcPacket, RpcHead) {
	rpcPacket := &RpcPacket{}
	proto.Unmarshal(buff, rpcPacket)
	if rpcPacket.RpcHead == nil {
		rpcPacket.RpcHead = &RpcHead{}
	}
	// actor funcname
	/*actorArgs := strings.Split(rpcPacket.FuncName, ".")
	if len(actorArgs) == 2 {
		rpcPacket.RpcHead.ActorName = actorArgs[0]
		rpcPacket.FuncName = actorArgs[1]
	} else {
		rpcPacket.FuncName = actorArgs[0]
	}*/

	return rpcPacket, *(*RpcHead)(rpcPacket.RpcHead)
}

//rpc Unmarshal
//pFuncType for  (this *X)func(conttext, params)
func UnmarshalBody(rpcPacket *RpcPacket, pFuncType reflect.Type) []interface{} {
	nCurLen := pFuncType.NumIn()
	params := make([]interface{}, nCurLen)
	buf := bytes.NewBuffer(rpcPacket.RpcBody)
	dec := gob.NewDecoder(buf)
	for i := 1; i < nCurLen; i++ {
		if i == 1 {
			params[1] = context.WithValue(context.Background(), "rpcHead", *(*RpcHead)(rpcPacket.RpcHead))
			continue
		}

		val := reflect.New(pFuncType.In(i))
		if i < int(rpcPacket.ArgLen+2) {
			dec.DecodeValue(val)
		}
		params[i] = val.Elem().Interface()
	}
	return params
}

func UnmarshalBodyCall(rpcPacket *RpcPacket, pFuncType reflect.Type) (error, []interface{}) {
	strErr := ""
	nCurLen := pFuncType.NumIn()
	params := make([]interface{}, nCurLen)
	buf := bytes.NewBuffer(rpcPacket.RpcBody)
	dec := gob.NewDecoder(buf)
	dec.Decode(&strErr)
	if strErr != "" {
		return errors.New(strErr), params
	}
	for i := 0; i < nCurLen; i++ {
		if i == 0 {
			params[0] = context.WithValue(context.Background(), "rpcHead", *(*RpcHead)(rpcPacket.RpcHead))
			continue
		}

		val := reflect.New(pFuncType.In(i))
		if i < int(rpcPacket.ArgLen+1) {
			dec.DecodeValue(val)
		}
		params[i] = val.Elem().Interface()
	}
	return nil, params
}

//rpc  UnmarshalPB
func unmarshalPB(bitstream *base.BitStream) (proto.Message, error) {
	packetName := bitstream.ReadString()
	nLen := bitstream.ReadInt(32)
	packetBuf := bitstream.ReadBits(nLen << 3)
	packet := reflect.New(proto.MessageType(packetName).Elem()).Interface().(proto.Message)
	err := proto.Unmarshal(packetBuf, packet)
	return packet, err
}
