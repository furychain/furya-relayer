package cosmos

import (
	"context"
	"fmt"
	"sync"

	rollapptypes "github.com/furychain/furya/x/rollapp/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	lock                       = &sync.Mutex{}
	furyaProviderSingleton *GridironSettlementProvider
)

type GridironSettlementProvider struct {
	*CosmosProvider
}

// NewSettlementProvider is creating a settlement provider which is a warrper for CosmosProvider
// and provides QueryLatestFinalizedHeight
func NewSettlementProvider(cp *CosmosProvider) (*GridironSettlementProvider, error) {
	lock.Lock()
	defer lock.Unlock()
	if furyaProviderSingleton != nil {
		return nil, fmt.Errorf("settlement was already initialized as %s. Cannot be initialized twich as %s",
			furyaProviderSingleton.ChainName(), cp.ChainName())
	}
	furyaProviderSingleton = &GridironSettlementProvider{cp}
	return furyaProviderSingleton, nil
}

// QueryLatestFinalizedHeight return the latest finalized height of a rollapp
func (cc *GridironSettlementProvider) QueryLatestFinalizedHeight(ctx context.Context, rollapId string) (int64, error) {
	qc := rollapptypes.NewQueryClient(cc)
	res, err := qc.LatestFinalizedStateInfo(ctx,
		&rollapptypes.QueryGetLatestFinalizedStateInfoRequest{RollappId: rollapId})

	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return -1, nil
		}
		return -1, err
	}
	if res == nil {
		return -1, fmt.Errorf("can't get latest-finalized-state info")
	}
	return int64(res.StateInfo.StartHeight + res.StateInfo.NumBlocks - 1), nil

}

func GetLatestFinalizedStateHeight(ctx context.Context, rollapId string) (int64, error) {
	return furyaProviderSingleton.QueryLatestFinalizedHeight(ctx, rollapId)
}
