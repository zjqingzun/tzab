package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterInterfaces(registrar codectypes.InterfaceRegistry) {
	registrar.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRecordAccess{},
	)

	registrar.RegisterImplementations((*sdk.Msg)(nil),
		&MsgUpdatePolicy{},
	)

	registrar.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRevokeAccess{},
	)

	registrar.RegisterImplementations((*sdk.Msg)(nil),
		&MsgGrantAccess{},
	)

	registrar.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRegisterResource{},
	)

	registrar.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRegisterIdentity{},
	)

	registrar.RegisterImplementations((*sdk.Msg)(nil),
		&MsgUpdateParams{},
	)
	msgservice.RegisterMsgServiceDesc(registrar, &_Msg_serviceDesc)
}
