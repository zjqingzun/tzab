package zt

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	ztsimulation "zerotrust/x/zt/simulation"
	"zerotrust/x/zt/types"
)

// GenerateGenesisState creates a randomized GenState of the module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	ztGenesis := types.GenesisState{
		Params: types.DefaultParams(),
	}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&ztGenesis)
}

// RegisterStoreDecoder registers a decoder.
func (am AppModule) RegisterStoreDecoder(_ simtypes.StoreDecoderRegistry) {}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	operations := make([]simtypes.WeightedOperation, 0)
	const (
		opWeightMsgRegisterIdentity          = "op_weight_msg_zt"
		defaultWeightMsgRegisterIdentity int = 100
	)

	var weightMsgRegisterIdentity int
	simState.AppParams.GetOrGenerate(opWeightMsgRegisterIdentity, &weightMsgRegisterIdentity, nil,
		func(_ *rand.Rand) {
			weightMsgRegisterIdentity = defaultWeightMsgRegisterIdentity
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgRegisterIdentity,
		ztsimulation.SimulateMsgRegisterIdentity(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))
	const (
		opWeightMsgRegisterResource          = "op_weight_msg_zt"
		defaultWeightMsgRegisterResource int = 100
	)

	var weightMsgRegisterResource int
	simState.AppParams.GetOrGenerate(opWeightMsgRegisterResource, &weightMsgRegisterResource, nil,
		func(_ *rand.Rand) {
			weightMsgRegisterResource = defaultWeightMsgRegisterResource
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgRegisterResource,
		ztsimulation.SimulateMsgRegisterResource(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))
	const (
		opWeightMsgGrantAccess          = "op_weight_msg_zt"
		defaultWeightMsgGrantAccess int = 100
	)

	var weightMsgGrantAccess int
	simState.AppParams.GetOrGenerate(opWeightMsgGrantAccess, &weightMsgGrantAccess, nil,
		func(_ *rand.Rand) {
			weightMsgGrantAccess = defaultWeightMsgGrantAccess
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgGrantAccess,
		ztsimulation.SimulateMsgGrantAccess(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))
	const (
		opWeightMsgRevokeAccess          = "op_weight_msg_zt"
		defaultWeightMsgRevokeAccess int = 100
	)

	var weightMsgRevokeAccess int
	simState.AppParams.GetOrGenerate(opWeightMsgRevokeAccess, &weightMsgRevokeAccess, nil,
		func(_ *rand.Rand) {
			weightMsgRevokeAccess = defaultWeightMsgRevokeAccess
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgRevokeAccess,
		ztsimulation.SimulateMsgRevokeAccess(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))
	const (
		opWeightMsgUpdatePolicy          = "op_weight_msg_zt"
		defaultWeightMsgUpdatePolicy int = 100
	)

	var weightMsgUpdatePolicy int
	simState.AppParams.GetOrGenerate(opWeightMsgUpdatePolicy, &weightMsgUpdatePolicy, nil,
		func(_ *rand.Rand) {
			weightMsgUpdatePolicy = defaultWeightMsgUpdatePolicy
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgUpdatePolicy,
		ztsimulation.SimulateMsgUpdatePolicy(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))
	const (
		opWeightMsgRecordAccess          = "op_weight_msg_zt"
		defaultWeightMsgRecordAccess int = 100
	)

	var weightMsgRecordAccess int
	simState.AppParams.GetOrGenerate(opWeightMsgRecordAccess, &weightMsgRecordAccess, nil,
		func(_ *rand.Rand) {
			weightMsgRecordAccess = defaultWeightMsgRecordAccess
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgRecordAccess,
		ztsimulation.SimulateMsgRecordAccess(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))

	return operations
}

// ProposalMsgs returns msgs used for governance proposals for simulations.
func (am AppModule) ProposalMsgs(simState module.SimulationState) []simtypes.WeightedProposalMsg {
	return []simtypes.WeightedProposalMsg{}
}
