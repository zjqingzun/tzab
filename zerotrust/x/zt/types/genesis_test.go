package types_test

import (
	"testing"

	"zerotrust/x/zt/types"

	"github.com/stretchr/testify/require"
)

func TestGenesisState_Validate(t *testing.T) {
	tests := []struct {
		desc     string
		genState *types.GenesisState
		valid    bool
	}{
		{
			desc:     "default is valid",
			genState: types.DefaultGenesis(),
			valid:    true,
		},
		{
			desc:     "valid genesis state",
			genState: &types.GenesisState{IdentityMap: []types.Identity{{Did: "0"}, {Did: "1"}}, ResourceMap: []types.Resource{{ResourceId: "0"}, {ResourceId: "1"}}, GrantMap: []types.Grant{{GrantId: "0"}, {GrantId: "1"}}, PolicyMap: []types.Policy{{PolicyId: "0"}, {PolicyId: "1"}}, AuditRecordList: []types.AuditRecord{{Id: 0}, {Id: 1}}, AuditRecordCount: 2}, valid: true,
		}, {
			desc: "duplicated identity",
			genState: &types.GenesisState{
				IdentityMap: []types.Identity{
					{
						Did: "0",
					},
					{
						Did: "0",
					},
				},
				ResourceMap: []types.Resource{{ResourceId: "0"}, {ResourceId: "1"}}, GrantMap: []types.Grant{{GrantId: "0"}, {GrantId: "1"}}, PolicyMap: []types.Policy{{PolicyId: "0"}, {PolicyId: "1"}}, AuditRecordList: []types.AuditRecord{{Id: 0}, {Id: 1}}, AuditRecordCount: 2,
			}, valid: false,
		}, {
			desc: "duplicated resource",
			genState: &types.GenesisState{
				ResourceMap: []types.Resource{
					{
						ResourceId: "0",
					},
					{
						ResourceId: "0",
					},
				},
				GrantMap: []types.Grant{{GrantId: "0"}, {GrantId: "1"}}, PolicyMap: []types.Policy{{PolicyId: "0"}, {PolicyId: "1"}}, AuditRecordList: []types.AuditRecord{{Id: 0}, {Id: 1}}, AuditRecordCount: 2,
			}, valid: false,
		}, {
			desc: "duplicated grant",
			genState: &types.GenesisState{
				GrantMap: []types.Grant{
					{
						GrantId: "0",
					},
					{
						GrantId: "0",
					},
				},
				PolicyMap: []types.Policy{{PolicyId: "0"}, {PolicyId: "1"}}, AuditRecordList: []types.AuditRecord{{Id: 0}, {Id: 1}}, AuditRecordCount: 2,
			}, valid: false,
		}, {
			desc: "duplicated policy",
			genState: &types.GenesisState{
				PolicyMap: []types.Policy{
					{
						PolicyId: "0",
					},
					{
						PolicyId: "0",
					},
				},
				AuditRecordList: []types.AuditRecord{{Id: 0}, {Id: 1}}, AuditRecordCount: 2,
			}, valid: false,
		}, {
			desc: "duplicated auditRecord",
			genState: &types.GenesisState{
				AuditRecordList: []types.AuditRecord{
					{
						Id: 0,
					},
					{
						Id: 0,
					},
				},
			},
			valid: false,
		}, {
			desc: "invalid auditRecord count",
			genState: &types.GenesisState{
				AuditRecordList: []types.AuditRecord{
					{
						Id: 1,
					},
				},
				AuditRecordCount: 0,
			},
			valid: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
