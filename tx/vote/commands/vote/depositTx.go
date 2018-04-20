package vote

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/lino-network/lino/client"
	"github.com/lino-network/lino/tx/vote"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
)

const (
	FlagAmount   = "amount"
	FlagUsername = "username"
)

// DepositVoterTxCmd will create a deposit tx and sign it with the given key
func DepositVoterTxCmd(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "voter-deposit",
		Short: "deposit money to be a voter",
		RunE:  sendDepositVoterTx(cdc),
	}
	cmd.Flags().String(FlagUsername, "", "deposit user")
	cmd.Flags().String(FlagAmount, "", "amount to deposit")
	return cmd
}

func sendDepositVoterTx(cdc *wire.Codec) client.CommandTxCallback {
	return func(cmd *cobra.Command, args []string) error {
		ctx := context.NewCoreContextFromViper()
		user := viper.GetString(FlagUsername)
		amount, err := sdk.NewRatFromDecimal(viper.GetString(FlagAmount))
		if err != nil {
			return err
		}

		// create the message
		msg := vote.NewVoterDepositMsg(user, amount)

		// build and sign the transaction, then broadcast to Tendermint
		res, signErr := ctx.SignBuildBroadcast(user, msg, cdc)

		if signErr != nil {
			return signErr
		}

		fmt.Printf("Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
		return nil
	}
}
