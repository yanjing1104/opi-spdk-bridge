// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2022 Dell Inc, or its subsidiaries.
// Copyright (C) 2023 Intel Corporation

// Package server implements the server
package server

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"

	"github.com/opiproject/opi-spdk-bridge/pkg/models"
)

// JSONRPC represents a structure to execute JSON RPC to SPDK
type JSONRPC struct {
	socket string
	id     uint64
}

// NewJSONRPC creates a new instance of JsonRpc
func NewJSONRPC(socketPath string) *JSONRPC {
	return &JSONRPC{
		socket: socketPath,
		id:     0,
	}
}

// RPCSock is unix domain socket to communicate with SPDK app or Vendor SDK
var RPCSock = flag.String("rpc_sock", "/var/tmp/spdk.sock", "Path to SPDK JSON RPC socket")

// DefaultJSONRPC is a default JSON RPC provider
var DefaultJSONRPC = NewJSONRPC("/var/tmp/spdk.sock")

// Call implements low level rpc request/response handling
func Call(method string, args, result interface{}) error {
	return DefaultJSONRPC.Call(method, args, result)
}

// Call implements low level rpc request/response handling
func (r *JSONRPC) Call(method string, args, result interface{}) error {
	id := atomic.AddUint64(&r.id, 1)
	request := models.RPCRequest{
		RPCVersion: models.JSONRPCVersion,
		ID:         id,
		Method:     method,
		Params:     args,
	}
	data, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("%s: %s", method, err)
	}

	log.Printf("Sending to SPDK: %s", data)

	// TODO: add also web option: resp, _ = webSocketCom(rpcClient, data)
	resp, _ := unixSocketCom(r.socket, data)

	var response models.RPCResponse
	err = json.NewDecoder(resp).Decode(&response)
	jsonresponse, _ := json.Marshal(response)
	log.Printf("Received from SPDK: %s", jsonresponse)
	if err != nil {
		return fmt.Errorf("%s: %s", method, err)
	}
	if response.ID != id {
		return fmt.Errorf("%s: json response ID mismatch", method)
	}
	if response.Error.Code != 0 {
		return fmt.Errorf("%s: json response error: %s", method, response.Error.Message)
	}
	err = json.Unmarshal(response.Result, &result)
	if err != nil {
		return fmt.Errorf("%s: %s", method, err)
	}
	return nil
}

func unixSocketCom(lrpcSock string, buf []byte) (io.Reader, error) {
	conn, err := net.Dial("unix", lrpcSock)
	if err != nil {
		log.Fatal(err)
	}
	_, err = conn.Write(buf)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	err = conn.(*net.UnixConn).CloseWrite()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return bufio.NewReader(conn), nil
}
