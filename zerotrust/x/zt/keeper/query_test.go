package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"zerotrust/x/zt/keeper"
	"zerotrust/x/zt/types"
)

func TestNewQueryServerImpl(t *testing.T) {
	f := initFixture(t)

	server := keeper.NewQueryServerImpl(f.keeper)

	require.NotNil(t, server)

	_, ok := server.(types.QueryServer)
	require.True(t, ok)
}
