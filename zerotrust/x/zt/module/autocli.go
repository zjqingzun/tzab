package zt

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	"zerotrust/x/zt/types"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: types.Query_serviceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Shows the parameters of the module",
				},
				{
					RpcMethod: "ListIdentity",
					Use:       "list-identity",
					Short:     "List all identity",
				},
				{
					RpcMethod:      "GetIdentity",
					Use:            "get-identity [id]",
					Short:          "Gets a identity",
					Alias:          []string{"show-identity"},
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "did"}},
				},
				{
					RpcMethod: "ListResource",
					Use:       "list-resource",
					Short:     "List all resource",
				},
				{
					RpcMethod:      "GetResource",
					Use:            "get-resource [id]",
					Short:          "Gets a resource",
					Alias:          []string{"show-resource"},
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "resource_id"}},
				},
				{
					RpcMethod: "ListGrant",
					Use:       "list-grant",
					Short:     "List all grant",
				},
				{
					RpcMethod:      "GetGrant",
					Use:            "get-grant [id]",
					Short:          "Gets a grant",
					Alias:          []string{"show-grant"},
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "grant_id"}},
				},
				{
					RpcMethod: "ListPolicy",
					Use:       "list-policy",
					Short:     "List all policy",
				},
				{
					RpcMethod:      "GetPolicy",
					Use:            "get-policy [id]",
					Short:          "Gets a policy",
					Alias:          []string{"show-policy"},
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "policy_id"}},
				},
				{
					RpcMethod: "ListAuditRecord",
					Use:       "list-audit-record",
					Short:     "List all auditRecord",
				},
				{
					RpcMethod:      "GetAuditRecord",
					Use:            "get-audit-record [id]",
					Short:          "Gets a auditRecord by id",
					Alias:          []string{"show-audit-record"},
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}},
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service:              types.Msg_serviceDesc.ServiceName,
			EnhanceCustomCommand: true, // only required if you want to use the custom command
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "UpdateParams",
					Skip:      true, // skipped because authority gated
				},
			},
		},
	}
}
