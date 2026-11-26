package keeper

import (
	"context"

	"zerotrust/x/zt/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx context.Context, genState types.GenesisState) error {
	for _, elem := range genState.IdentityMap {
		if err := k.Identity.Set(ctx, elem.Did, elem); err != nil {
			return err
		}
	}
	for _, elem := range genState.ResourceMap {
		if err := k.Resource.Set(ctx, elem.ResourceId, elem); err != nil {
			return err
		}
	}
	for _, elem := range genState.GrantMap {
		if err := k.Grant.Set(ctx, elem.GrantId, elem); err != nil {
			return err
		}
	}
	for _, elem := range genState.PolicyMap {
		if err := k.Policy.Set(ctx, elem.PolicyId, elem); err != nil {
			return err
		}
	}
	for _, elem := range genState.AuditRecordList {
		if err := k.AuditRecord.Set(ctx, elem.Id, elem); err != nil {
			return err
		}
	}

	if err := k.AuditRecordSeq.Set(ctx, genState.AuditRecordCount); err != nil {
		return err
	}

	return k.Params.Set(ctx, genState.Params)
}

// ExportGenesis returns the module's exported genesis.
func (k Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	var err error

	genesis := types.DefaultGenesis()
	genesis.Params, err = k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	if err := k.Identity.Walk(ctx, nil, func(_ string, val types.Identity) (stop bool, err error) {
		genesis.IdentityMap = append(genesis.IdentityMap, val)
		return false, nil
	}); err != nil {
		return nil, err
	}
	if err := k.Resource.Walk(ctx, nil, func(_ string, val types.Resource) (stop bool, err error) {
		genesis.ResourceMap = append(genesis.ResourceMap, val)
		return false, nil
	}); err != nil {
		return nil, err
	}
	if err := k.Grant.Walk(ctx, nil, func(_ string, val types.Grant) (stop bool, err error) {
		genesis.GrantMap = append(genesis.GrantMap, val)
		return false, nil
	}); err != nil {
		return nil, err
	}
	if err := k.Policy.Walk(ctx, nil, func(_ string, val types.Policy) (stop bool, err error) {
		genesis.PolicyMap = append(genesis.PolicyMap, val)
		return false, nil
	}); err != nil {
		return nil, err
	}
	err = k.AuditRecord.Walk(ctx, nil, func(key uint64, elem types.AuditRecord) (bool, error) {
		genesis.AuditRecordList = append(genesis.AuditRecordList, elem)
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	genesis.AuditRecordCount, err = k.AuditRecordSeq.Peek(ctx)
	if err != nil {
		return nil, err
	}

	return genesis, nil
}
