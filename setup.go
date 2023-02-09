package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
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

func setupRpcOracles() (L1PreimageOracle, L2PreimageOracle, error) {
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
	store, err := NewDiskStore(storePath)
	if err != nil {
		return nil, nil, fmt.Errorf("creating disk store: %w", err)
	}

	l1Oracle := NewLoadingL1Chain(l1Client, store)
	l2Oracle := NewLoadingL2Chain(rpcClient, BlockStore{store})
	return l1Oracle, l2Oracle, nil
}