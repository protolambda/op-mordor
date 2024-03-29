package main

import (
	"context"
	"fmt"
	"op-mordor/l1"
	"op-mordor/l2"
	"op-mordor/oracle"
	"op-mordor/store"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
	_ "github.com/joho/godotenv/autoload"
)

const dialTimeout = 5 * time.Second // TODO: flag or env

var (
	l1RpcURL  string
	l2RpcURL  string
	storePath = "/tmp/mordor"
)

func setupEnv() {
	l1RpcURL = os.Getenv("OP_L1_RPC")
	l2RpcURL = os.Getenv("OP_L2_RPC")
	if path := os.Getenv("OP_STORE_PATH"); path != "" {
		storePath = path
	}
}

func setupRpcOracles(logger log.Logger) (oracle.L1Oracle, oracle.L2Oracle, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dialTimeout)
	defer cancel()

	l1Client, err := ethclient.DialContext(ctx, l1RpcURL)
	if err != nil {
		return nil, nil, fmt.Errorf("dialing l1 rpc: %w", err)
	}
	rpcClient, err := rpc.DialContext(ctx, l2RpcURL)
	if err != nil {
		return nil, nil, fmt.Errorf("dialing l2 rpc: %w", err)
	}
	dstore, err := store.NewDiskStore(storePath)
	if err != nil {
		return nil, nil, fmt.Errorf("creating disk store: %w", err)
	}

	l1Oracle := l1.NewLoadingL1Chain(logger, l1Client, dstore)
	l2Oracle := l2.NewLoadingL2Chain(logger, rpcClient, dstore, dstore)
	return l1Oracle, l2Oracle, nil
}
