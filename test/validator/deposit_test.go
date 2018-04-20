package validator

import (
	"testing"
	"time"

	"github.com/lino-network/lino/test"
	val "github.com/lino-network/lino/tx/validator"
	vote "github.com/lino-network/lino/tx/vote"
	"github.com/lino-network/lino/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	crypto "github.com/tendermint/go-crypto"
)

// test normal transfer to account name
func TestValidatorDeposit(t *testing.T) {
	newAccountPriv := crypto.GenPrivKeyEd25519()
	newAccountName := "newUser"
	newValidatorPriv := crypto.GenPrivKeyEd25519()

	baseTime := time.Now().Unix() + 100
	lb := test.NewTestLinoBlockchain(t, test.DefaultNumOfVal)

	test.CreateAccount(t, newAccountName, lb, 0, newAccountPriv, 5000)

	voteDepositMsg := vote.NewVoterDepositMsg(newAccountName, types.LNO(sdk.NewRat(3000)))
	test.SignCheckDeliver(t, lb, voteDepositMsg, 0, true, newAccountPriv, baseTime)

	valDepositMsg := val.NewValidatorDepositMsg(
		newAccountName, types.LNO(sdk.NewRat(1000)), newValidatorPriv.PubKey())
	test.SignCheckDeliver(t, lb, valDepositMsg, 1, true, newAccountPriv, baseTime)
	test.CheckAllValidatorList(t, newAccountName, true, lb)

	valDepositMsg = val.NewValidatorDepositMsg(
		newAccountName, types.LNO(sdk.NewRat(1)), newValidatorPriv.PubKey())
	test.SignCheckDeliver(t, lb, valDepositMsg, 2, true, newAccountPriv, baseTime)
	test.CheckOncallValidatorList(t, newAccountName, true, lb)
	test.CheckAllValidatorList(t, newAccountName, true, lb)
}