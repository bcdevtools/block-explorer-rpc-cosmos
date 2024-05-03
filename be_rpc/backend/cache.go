package backend

import (
	"context"
	"fmt"
	berpcutils "github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/utils"
	"github.com/cosmos/cosmos-sdk/codec"
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

type tendermintValidatorsCache struct {
	cacheController *baseCacheController
	validators      []*tmtypes.Validator
	tmClient        client.Client
}

const validatorsCacheExpiration = 100

func NewTendermintValidatorsCache(tmClient client.Client) *tendermintValidatorsCache {
	funcIsExpired := func(expirationAnchor, valueToCompare any) bool {
		return valueToCompare.(int64) > expirationAnchor.(int64)
	}
	return &tendermintValidatorsCache{
		cacheController: NewBaseCacheController(funcIsExpired),
		tmClient:        tmClient,
	}
}

func (vc *tendermintValidatorsCache) GetValidators() (vals []*tmtypes.Validator, err error) {
	isExpired, errCheckExpired := vc.IsCacheExpired()
	if errCheckExpired != nil {
		err = errCheckExpired
		return
	}

	if !isExpired {
		return vc.validators, nil
	}

	vc.cacheController.rwMutex.Lock()
	defer vc.cacheController.rwMutex.Unlock()

	isExpired, height, errCheckExpired := vc.isCacheExpired(false)
	if errCheckExpired != nil {
		err = errCheckExpired
		return
	}
	if !isExpired { // prevent race condition by re-checking after acquiring the lock
		return vc.validators, nil
	}

	errReloadCache := vc.reloadCacheWithoutLock(height)
	if errReloadCache != nil {
		err = errReloadCache
		return
	}

	return vc.validators, nil
}

func (vc *tendermintValidatorsCache) IsCacheExpired() (expired bool, err error) {
	expired, _, err = vc.isCacheExpired(true)
	return
}

func (vc *tendermintValidatorsCache) isCacheExpired(lock bool) (expired bool, latestHeight int64, err error) {
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

// reloadCacheWithoutLock performs reload cache. Lock acquire must be performed before calling this.
func (vc *tendermintValidatorsCache) reloadCacheWithoutLock(height int64) error {
	var page = 1
	var perPage = 200

	resValidators, err := vc.tmClient.Validators(context.Background(), &height, &page, &perPage)
	if err != nil {
		return err
	}

	vc.validators = resValidators.Validators
	vc.cacheController.UpdateExpirationAnchor(height + validatorsCacheExpiration)

	return nil
}

type validatorsConsAddrToValAddr struct {
	cacheController             *baseCacheController
	validatorsConsAddrToValAddr map[string]string
	tmClient                    client.Client
	stakingQueryClient          stakingtypes.QueryClient
	codec                       codec.Codec
}

func NewValidatorsConsAddrToValAddrCache(tmClient client.Client, stakingQueryClient stakingtypes.QueryClient, codec codec.Codec) *validatorsConsAddrToValAddr {
	funcIsExpired := func(expirationAnchor, valueToCompare any) bool {
		return valueToCompare.(int64) > expirationAnchor.(int64)
	}
	return &validatorsConsAddrToValAddr{
		cacheController:             NewBaseCacheController(funcIsExpired),
		validatorsConsAddrToValAddr: make(map[string]string),
		tmClient:                    tmClient,
		stakingQueryClient:          stakingQueryClient,
		codec:                       codec,
	}
}

func (vc *validatorsConsAddrToValAddr) GetValAddrFromConsAddr(consAddr string) (valAddr string, found bool, err error) {
	isExpired, errCheckExpired := vc.IsCacheExpired()
	if errCheckExpired != nil {
		err = errCheckExpired
		return
	}

	if !isExpired {
		valAddr, found = vc.validatorsConsAddrToValAddr[consAddr]
		return
	}

	vc.cacheController.rwMutex.Lock()
	defer vc.cacheController.rwMutex.Unlock()

	isExpired, height, errCheckExpired := vc.isCacheExpired(false)
	if errCheckExpired != nil {
		err = errCheckExpired
		return
	}
	if !isExpired { // prevent race condition by re-checking after acquiring the lock
		valAddr, found = vc.validatorsConsAddrToValAddr[consAddr]
		return
	}

	errReloadCache := vc.reloadCacheWithoutLock(height)
	if errReloadCache != nil {
		err = errReloadCache
		return
	}

	valAddr, found = vc.validatorsConsAddrToValAddr[consAddr]
	return
}

func (vc *validatorsConsAddrToValAddr) GetValAddrAndConsAddr(consOrValAddr string) (valAddr, consAddr string, found bool, err error) {
	isExpired, errCheckExpired := vc.IsCacheExpired()
	if errCheckExpired != nil {
		err = errCheckExpired
		return
	}

	lookupUnexpiredData := func() (valAddr, consAddr string, found bool) {
		for consAddr, valAddr := range vc.validatorsConsAddrToValAddr {
			if consAddr == consOrValAddr || valAddr == consOrValAddr {
				return valAddr, consAddr, true
			}
		}

		return "", "", false
	}

	if !isExpired {
		valAddr, consAddr, found = lookupUnexpiredData()
		return
	}

	vc.cacheController.rwMutex.Lock()
	defer vc.cacheController.rwMutex.Unlock()

	isExpired, height, errCheckExpired := vc.isCacheExpired(false)
	if errCheckExpired != nil {
		err = errCheckExpired
		return
	}
	if !isExpired { // prevent race condition by re-checking after acquiring the lock
		valAddr, consAddr, found = lookupUnexpiredData()
		return
	}

	errReloadCache := vc.reloadCacheWithoutLock(height)
	if errReloadCache != nil {
		err = errReloadCache
		return
	}

	valAddr, consAddr, found = lookupUnexpiredData()
	return
}

func (vc *validatorsConsAddrToValAddr) IsCacheExpired() (expired bool, err error) {
	expired, _, err = vc.isCacheExpired(true)
	return
}

func (vc *validatorsConsAddrToValAddr) isCacheExpired(lock bool) (expired bool, latestHeight int64, err error) {
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

// reloadCacheWithoutLock performs reload cache. Lock acquire must be performed before calling this.
func (vc *validatorsConsAddrToValAddr) reloadCacheWithoutLock(height int64) error {
	var perPage = 200

	stakingVals, errStakingVals := vc.stakingQueryClient.Validators(context.Background(), &stakingtypes.QueryValidatorsRequest{
		Pagination: &query.PageRequest{
			Offset: 0,
			Limit:  uint64(perPage),
		},
	})
	if errStakingVals != nil {
		return errStakingVals
	}

	for _, val := range stakingVals.Validators {
		consAddr, success := berpcutils.FromAnyPubKeyToConsensusAddress(val.ConsensusPubkey, vc.codec)
		if !success {
			panic(fmt.Sprintf("failed to convert consensus address of type %s which map to %s", val.ConsensusPubkey.TypeUrl, val.OperatorAddress))
			//continue
		}

		consAddrStr := consAddr.String()
		vc.validatorsConsAddrToValAddr[consAddrStr] = val.OperatorAddress
	}

	vc.cacheController.UpdateExpirationAnchor(height + validatorsCacheExpiration)

	return nil
}
