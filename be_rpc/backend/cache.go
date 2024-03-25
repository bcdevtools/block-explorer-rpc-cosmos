package backend

import (
	"context"
	"fmt"
	berpcutils "github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/tendermint/tendermint/rpc/client"
	tmtypes "github.com/tendermint/tendermint/types"
	"sync"
)

type baseCacheController struct {
	rwMutex          *sync.RWMutex
	expirationAnchor any
	funcIsExpired    func(expirationAnchor, valueToCompare any) bool
}

func NewBaseCacheController(expirationChecker func(expirationAnchor, valueToCompare any) bool) *baseCacheController {
	return &baseCacheController{
		rwMutex:       &sync.RWMutex{},
		funcIsExpired: expirationChecker,
	}
}

func (bcc *baseCacheController) UpdateExpirationAnchor(expirationAnchor any) {
	bcc.expirationAnchor = expirationAnchor
}

func (bcc *baseCacheController) IsExpired(valueToCompare any) bool {
	if bcc.expirationAnchor == nil {
		return true
	}

	return bcc.funcIsExpired(bcc.expirationAnchor, valueToCompare)
}

type validatorsCache struct {
	cacheController             *baseCacheController
	validators                  []*tmtypes.Validator
	validatorsConsAddrToValAddr map[string]string
	tmClient                    client.Client
	stakingQueryClient          stakingtypes.QueryClient
	codec                       codec.Codec
}

const validatorsCacheExpiration = 100

func NewValidatorsCache(tmClient client.Client, stakingQueryClient stakingtypes.QueryClient, codec codec.Codec) *validatorsCache {
	funcIsExpired := func(expirationAnchor, valueToCompare any) bool {
		return valueToCompare.(int64) > expirationAnchor.(int64)+validatorsCacheExpiration
	}
	return &validatorsCache{
		cacheController:             NewBaseCacheController(funcIsExpired),
		validatorsConsAddrToValAddr: make(map[string]string),
		tmClient:                    tmClient,
		stakingQueryClient:          stakingQueryClient,
		codec:                       codec,
	}
}

func (vc *validatorsCache) GetValidators() (vals []*tmtypes.Validator, err error) {
	isExpired, err := vc.IsCacheExpired()
	if err != nil {
		return nil, err
	}

	if !isExpired {
		return vc.validators, nil
	}

	vc.cacheController.rwMutex.Lock()
	defer vc.cacheController.rwMutex.Unlock()

	isExpired, height, err := vc.isCacheExpired(false)
	if !isExpired { // prevent race condition by re-checking after acquiring the lock
		return vc.validators, nil
	}

	var page = 1
	var perPage = 200

	resValidators, err := vc.tmClient.Validators(context.Background(), &height, &page, &perPage)
	if err != nil {
		return nil, err
	}

	validators := resValidators.Validators

	var missingAnyValidatorAddr bool
	for _, validator := range validators {
		consAddrStr := sdk.ConsAddress(validator.Address).String()
		_, foundValAddr := vc.validatorsConsAddrToValAddr[consAddrStr]
		if foundValAddr {
			continue
		}

		missingAnyValidatorAddr = true
		break
	}

	if missingAnyValidatorAddr {
		stakingVals, err := vc.stakingQueryClient.Validators(context.Background(), &stakingtypes.QueryValidatorsRequest{
			Pagination: &query.PageRequest{
				Offset: 0,
				Limit:  uint64(perPage),
			},
		})
		if err != nil {
			return nil, err
		}

		for _, val := range stakingVals.Validators {
			consAddr, success := berpcutils.FromAnyPubKeyToConsensusAddress(val.ConsensusPubkey, vc.codec)
			if !success {
				fmt.Println("Failed to convert consensus address")
				continue
			}

			consAddrStr := consAddr.String()
			vc.validatorsConsAddrToValAddr[consAddrStr] = val.OperatorAddress
		}
	}

	vc.validators = validators
	vc.cacheController.UpdateExpirationAnchor(height)

	return validators, nil
}

func (vc *validatorsCache) IsCacheExpired() (expired bool, err error) {
	expired, _, err = vc.isCacheExpired(true)
	return
}

func (vc *validatorsCache) GetValAddress(consAddr string) (valAddr string, found bool) {
	vc.cacheController.rwMutex.RLock()
	defer vc.cacheController.rwMutex.RUnlock()

	valAddr, found = vc.validatorsConsAddrToValAddr[consAddr]
	return
}

func (vc *validatorsCache) isCacheExpired(lock bool) (expired bool, latestHeight int64, err error) {
	resStatus, err := vc.tmClient.Status(context.Background())
	if err != nil {
		return false, 0, err
	}

	if lock {
		vc.cacheController.rwMutex.Lock()
		defer vc.cacheController.rwMutex.Unlock()
	}

	latestHeight = resStatus.SyncInfo.LatestBlockHeight
	expired = vc.cacheController.IsExpired(latestHeight)
	return
}
