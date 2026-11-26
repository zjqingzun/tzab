package keeper_test

import (
	"testing"

	"zerotrust/x/zt/types"

	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params:      types.DefaultParams(),
		IdentityMap: []types.Identity{{Did: "0"}, {Did: "1"}}, ResourceMap: []types.Resource{{ResourceId: "0"}, {ResourceId: "1"}}, GrantMap: []types.Grant{{GrantId: "0"}, {GrantId: "1"}}, PolicyMap: []types.Policy{{PolicyId: "0"}, {PolicyId: "1"}}, AuditRecordList: []types.AuditRecord{{Id: 0}, {Id: 1}},
		AuditRecordCount: 2,
	}
	f := initFixture(t)
	err := f.keeper.InitGenesis(f.ctx, genesisState)
	require.NoError(t, err)
	got, err := f.keeper.ExportGenesis(f.ctx)
	require.NoError(t, err)
	require.NotNil(t, got)

	require.EqualExportedValues(t, genesisState.Params, got.Params)
	require.EqualExportedValues(t, genesisState.IdentityMap, got.IdentityMap)
	require.EqualExportedValues(t, genesisState.ResourceMap, got.ResourceMap)
	require.EqualExportedValues(t, genesisState.GrantMap, got.GrantMap)
	require.EqualExportedValues(t, genesisState.PolicyMap, got.PolicyMap)
	require.EqualExportedValues(t, genesisState.AuditRecordList, got.AuditRecordList)
	require.Equal(t, genesisState.AuditRecordCount, got.AuditRecordCount)

}
