package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	corestore "cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"

	"zerotrust/x/zt/types"
)

type Keeper struct {
	storeService corestore.KVStoreService
	cdc          codec.Codec
	addressCodec address.Codec
	// Address capable of executing a MsgUpdateParams message.
	// Typically, this should be the x/gov module account.
	authority []byte

	Schema         collections.Schema
	Params         collections.Item[types.Params]
	Identity       collections.Map[string, types.Identity]
	Resource       collections.Map[string, types.Resource]
	Grant          collections.Map[string, types.Grant]
	Policy         collections.Map[string, types.Policy]
	AuditRecordSeq collections.Sequence
	AuditRecord    collections.Map[uint64, types.AuditRecord]
}

func NewKeeper(
	storeService corestore.KVStoreService,
	cdc codec.Codec,
	addressCodec address.Codec,
	authority []byte,

) Keeper {
	if _, err := addressCodec.BytesToString(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address %s: %s", authority, err))
	}

	sb := collections.NewSchemaBuilder(storeService)

	k := Keeper{
		storeService: storeService,
		cdc:          cdc,
		addressCodec: addressCodec,
		authority:    authority,

		Params:   collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		Identity: collections.NewMap(sb, types.IdentityKey, "identity", collections.StringKey, codec.CollValue[types.Identity](cdc)), Resource: collections.NewMap(sb, types.ResourceKey, "resource", collections.StringKey, codec.CollValue[types.Resource](cdc)), Grant: collections.NewMap(sb, types.GrantKey, "grant", collections.StringKey, codec.CollValue[types.Grant](cdc)), Policy: collections.NewMap(sb, types.PolicyKey, "policy", collections.StringKey, codec.CollValue[types.Policy](cdc)), AuditRecord: collections.NewMap(sb, types.AuditRecordKey, "auditRecord", collections.Uint64Key, codec.CollValue[types.AuditRecord](cdc)),
		AuditRecordSeq: collections.NewSequence(sb, types.AuditRecordCountKey, "auditRecordSequence"),
	}
	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema

	return k
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() []byte {
	return k.authority
}
