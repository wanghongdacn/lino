package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/lino-network/lino/client"
	post "github.com/lino-network/lino/tx/post"

	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/lino-network/lino/types"
)

// ViewTxCmd will create a view tx and sign it with the given key
func ViewTxCmd(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view",
		Short: "view a post",
		RunE:  sendViewTx(cdc),
	}
	cmd.Flags().String(client.FlagUser, "", "view user of this transaction")
	cmd.Flags().String(client.FlagPostID, "", "post id to identify this post for the author")
	cmd.Flags().String(client.FlagAuthor, "", "title for the post")
	return cmd
}

// send view transaction to the blockchain
func sendViewTx(cdc *wire.Codec) client.CommandTxCallback {
	return func(cmd *cobra.Command, args []string) error {
		ctx := client.NewCoreContextFromViper()
		username := viper.GetString(client.FlagUser)
		author := viper.GetString(client.FlagAuthor)
		postID := viper.GetString(client.FlagPostID)

		msg := post.NewViewMsg(types.AccountKey(username), types.AccountKey(author), postID)

		// build and sign the transaction, then broadcast to Tendermint
		res, err := ctx.SignBuildBroadcast(msg, cdc)

		if err != nil {
			return err
		}

		fmt.Printf("Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
		return nil
	}
}
