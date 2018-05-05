package model

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/lino-network/lino/types"
)

var (
	ProposalSubstore     = []byte{0x00}
	ProposalListSubStore = []byte{0x01}
)

type ProposalStorage struct {
	key sdk.StoreKey
	cdc *wire.Codec
}

func NewProposalStorage(key sdk.StoreKey) ProposalStorage {
	cdc := wire.NewCodec()
	wire.RegisterCrypto(cdc)
	vs := ProposalStorage{
		key: key,
		cdc: cdc,
	}
	return vs
}

func (ps ProposalStorage) InitGenesis(ctx sdk.Context) sdk.Error {
	proposalLst := &ProposalList{}
	if err := ps.SetProposalList(ctx, proposalLst); err != nil {
		return err
	}
	return nil
}

func (ps ProposalStorage) GetProposalList(ctx sdk.Context) (*ProposalList, sdk.Error) {
	store := ctx.KVStore(ps.key)
	lstByte := store.Get(GetProposalListKey())
	if lstByte == nil {
		return nil, ErrGetProposal()
	}
	lst := new(ProposalList)
	if err := ps.cdc.UnmarshalJSON(lstByte, lst); err != nil {
		return nil, ErrProposalUnmarshalError(err)
	}
	return lst, nil
}

func (ps ProposalStorage) SetProposalList(ctx sdk.Context, lst *ProposalList) sdk.Error {
	store := ctx.KVStore(ps.key)
	lstByte, err := ps.cdc.MarshalJSON(*lst)
	if err != nil {
		return ErrProposalMarshalError(err)
	}
	store.Set(GetProposalListKey(), lstByte)
	return nil
}

// onle support change parameter proposal now
func (ps ProposalStorage) GetProposal(ctx sdk.Context, proposalID types.ProposalKey) (*ChangeParameterProposal, sdk.Error) {
	store := ctx.KVStore(ps.key)
	proposalByte := store.Get(GetProposalKey(proposalID))
	if proposalByte == nil {
		return nil, ErrGetProposal()
	}
	proposal := new(ChangeParameterProposal)
	if err := ps.cdc.UnmarshalJSON(proposalByte, proposal); err != nil {
		return nil, ErrProposalUnmarshalError(err)
	}
	return proposal, nil
}

// onle support change parameter proposal now
func (ps ProposalStorage) SetProposal(ctx sdk.Context, proposalID types.ProposalKey, proposal *ChangeParameterProposal) sdk.Error {
	store := ctx.KVStore(ps.key)
	proposalByte, err := ps.cdc.MarshalJSON(*proposal)
	if err != nil {
		return ErrProposalMarshalError(err)
	}
	store.Set(GetProposalKey(proposalID), proposalByte)
	return nil
}

func (ps ProposalStorage) DeleteProposal(ctx sdk.Context, proposalID types.ProposalKey) sdk.Error {
	store := ctx.KVStore(ps.key)
	store.Delete(GetProposalKey(proposalID))
	return nil
}

func GetProposalKey(proposalID types.ProposalKey) []byte {
	return append(ProposalSubstore, proposalID...)
}

func GetProposalListKey() []byte {
	return ProposalListSubStore
}
