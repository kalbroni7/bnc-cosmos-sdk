package ibc

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// IBC Keeper
type Keeper struct {
	storeKey  sdk.StoreKey
	codespace sdk.CodespaceType
}

func NewKeeper(storeKey sdk.StoreKey, codespace sdk.CodespaceType) Keeper {
	return Keeper{
		storeKey:  storeKey,
		codespace: codespace,
	}
}

func (k Keeper) CreateIBCPackage(ctx sdk.Context, destChainID sdk.CrossChainID, channelName string, value []byte) sdk.Error {
	channelID, err := sdk.GetChannelID(channelName)
	if err != nil {
		return sdk.ErrInternal(err.Error())
	}
	sequence := k.GetNextSequence(ctx, destChainID, channelName)
	key := buildIBCPackageKey(sdk.GetSourceChainID(), destChainID, channelID, sequence)
	kvStore := ctx.KVStore(k.storeKey)
	if kvStore.Has(key) {
		return ErrDuplicatedSequence(DefaultCodespace, "duplicated sequence")
	}
	kvStore.Set(key, value)
	k.incrSequence(ctx, destChainID, channelID)
	return nil
}

func (k *Keeper) GetIBCPackage(ctx sdk.Context, destChainID sdk.CrossChainID, channelName string, sequence uint64) ([]byte, error) {
	channelID, err := sdk.GetChannelID(channelName)
	if err != nil {
		return nil, err
	}
	kvStore := ctx.KVStore(k.storeKey)
	key := buildIBCPackageKey(sdk.GetSourceChainID(), destChainID, channelID, sequence)
	return kvStore.Get(key), nil
}

func (k Keeper) CleanupIBCPackage(ctx sdk.Context, destChainID sdk.CrossChainID, channelName string, confirmedSequence uint64) {
	channelID, err := sdk.GetChannelID(channelName)
	if err != nil {
		return
	}
	prefixKey := buildIBCPackageKeyPrefix(sdk.GetSourceChainID(), destChainID, channelID)
	kvStore := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(kvStore, prefixKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		packageKey := iterator.Key()
		if len(packageKey) != prefixLength+sourceChainIDLength+destChainIDLength+channelIDLength+sequenceLength {
			continue
		}
		sequence := binary.BigEndian.Uint64(packageKey[prefixLength+sourceChainIDLength+destChainIDLength+channelIDLength:])
		if sequence > confirmedSequence {
			break
		}
		kvStore.Delete(packageKey)
	}
}

func (k *Keeper) GetNextSequence(ctx sdk.Context, destChainID sdk.CrossChainID, channelName string) uint64 {
	channelID, err := sdk.GetChannelID(channelName)
	if err != nil {
		return 0
	}

	kvStore := ctx.KVStore(k.storeKey)
	bz := kvStore.Get(buildChannelSequenceKey(destChainID, channelID))
	if bz == nil {
		return 0
	}
	return binary.BigEndian.Uint64(bz)
}

func (k *Keeper) incrSequence(ctx sdk.Context, destChainID sdk.CrossChainID, channelID sdk.ChannelID) {
	var sequence uint64
	kvStore := ctx.KVStore(k.storeKey)
	bz := kvStore.Get(buildChannelSequenceKey(destChainID, channelID))
	if bz == nil {
		sequence = 0
	} else {
		sequence = binary.BigEndian.Uint64(bz)
	}

	sequenceBytes := make([]byte, sequenceLength)
	binary.BigEndian.PutUint64(sequenceBytes, sequence+1)
	kvStore.Set(buildChannelSequenceKey(destChainID, channelID), sequenceBytes)
}
