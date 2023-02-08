package main

type L2BlockFetcher struct {
	oracle   *PreimageOracle
	fallback interface{} // TODO RPC endpoint
}

func NewL2BlockFetcher(oracle *PreimageOracle) *L2BlockFetcher {
	// TODO rpc
	return &L2BlockFetcher{
		oracle:   oracle,
		fallback: nil,
	}
}
