package types

import (
	"bytes"
	"encoding/json"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire"
	"github.com/tendermint/go-wire/data"
	. "github.com/tendermint/tmlibs/common"
)

/*
Tx (Transaction) is an atomic operation on the ledger state.

Account Types:
 - SendTx         Send coins to address
 - AppTx         Send a msg to a contract that runs in the vm
*/
type Tx interface {
	AssertIsTx()
	SignBytes(chainID string) []byte
}

// Types of Tx implementations
const (
	// Account transactions
	TxTypeSend = byte(0x01)
	TxTypeApp  = byte(0x02)
	TxTypePost  = byte(0x03)
	TxTypeLike  = byte(0x71)
	TxTypeDonate = byte(0x80)
	TxNameSend = "send"
	TxNameApp  = "app"
	TxNamePost  = "post"
	TxNameDonate = "donate"
	TxNameLike = "like"
)

func (_ *SendTx) AssertIsTx() {}
func (_ *AppTx) AssertIsTx()  {}
func (_ *PostTx) AssertIsTx()  {}
func (_ *DonateTx) AssertIsTx() {}
func (_ *LikeTx) AssertIsTx()  {}

var txMapper data.Mapper

// register both private key types with go-wire/data (and thus go-wire)
func init() {
	txMapper = data.NewMapper(TxS{}).
		RegisterImplementation(&SendTx{}, TxNameSend, TxTypeSend).
		RegisterImplementation(&AppTx{}, TxNameApp, TxTypeApp).
		RegisterImplementation(&PostTx{}, TxNamePost, TxTypePost).
		RegisterImplementation(&DonateTx{}, TxNameDonate, TxTypeDonate).
		RegisterImplementation(&LikeTx{}, TxNameLike, TxTypeLike)
}

// TxS add json serialization to Tx
type TxS struct {
	Tx `json:"unwrap"`
}

func (p TxS) MarshalJSON() ([]byte, error) {
	return txMapper.ToJSON(p.Tx)
}

func (p *TxS) UnmarshalJSON(data []byte) (err error) {
	parsed, err := txMapper.FromJSON(data)
	if err == nil {
		p.Tx = parsed.(Tx)
	}
	return
}

//-----------------------------------------------------------------------------

type TxInput struct {
	Address   data.Bytes       `json:"address"`   // Hash of the PubKey
	Coins     Coins            `json:"coins"`     //
	Sequence  int              `json:"sequence"`  // Must be 1 greater than the last committed TxInput
	Signature crypto.Signature `json:"signature"` // Depends on the PubKey type and the whole Tx
	PubKey    crypto.PubKey    `json:"pub_key"`   // Is present iff Sequence == 0
}

func (txIn TxInput) ValidateBasic() abci.Result {
	if len(txIn.Address) != 20 {
		return abci.ErrBaseInvalidInput.AppendLog("Invalid address length")
	}
	if !txIn.Coins.IsValid() {
		return abci.ErrBaseInvalidInput.AppendLog(Fmt("Invalid coins %v", txIn.Coins))
	}
	if txIn.Coins.IsZero() {
		return abci.ErrBaseInvalidInput.AppendLog("Coins cannot be zero")
	}
	if txIn.Sequence <= 0 {
		return abci.ErrBaseInvalidInput.AppendLog("Sequence must be greater than 0")
	}
	if txIn.Sequence == 1 && txIn.PubKey.Empty() {
		return abci.ErrBaseInvalidInput.AppendLog("PubKey must be present when Sequence == 1")
	}
	if txIn.Sequence > 1 && !txIn.PubKey.Empty() {
		return abci.ErrBaseInvalidInput.AppendLog("PubKey must be nil when Sequence > 1")
	}
	return abci.OK
}

func (txIn TxInput) String() string {
	return Fmt("TxInput{%X,%v,%v,%v,%v}", txIn.Address, txIn.Coins, txIn.Sequence, txIn.Signature, txIn.PubKey)
}

func NewTxInput(pubKey crypto.PubKey, coins Coins, sequence int) TxInput {
	input := TxInput{
		Address:  pubKey.Address(),
		Coins:    coins,
		Sequence: sequence,
	}
	if sequence == 1 {
		input.PubKey = pubKey
	}
	return input
}

//-----------------------------------------------------------------------------

type TxOutput struct {
	Address data.Bytes `json:"address"` // Hash of the PubKey
	Coins   Coins      `json:"coins"`   //
}

// An output destined for another chain may be formatted as `chainID/address`.
// ChainAndAddress returns the chainID prefix and the address.
// If there is no chainID prefix, the first returned value is nil.
func (txOut TxOutput) ChainAndAddress() ([]byte, []byte, abci.Result) {
	var chainPrefix []byte
	address := txOut.Address
	if len(address) > 20 {
		spl := bytes.SplitN(address, []byte("/"), 2)
		if len(spl) != 2 {
			return nil, nil, abci.ErrBaseInvalidOutput.AppendLog("Invalid address format")
		}
		chainPrefix = spl[0]
		address = spl[1]
	}

	if len(address) != 20 {
		return nil, nil, abci.ErrBaseInvalidOutput.AppendLog("Invalid address length")
	}
	return chainPrefix, address, abci.OK
}

func (txOut TxOutput) ValidateBasic() abci.Result {
	_, _, r := txOut.ChainAndAddress()
	if r.IsErr() {
		return r
	}

	if !txOut.Coins.IsValid() {
		return abci.ErrBaseInvalidOutput.AppendLog(Fmt("Invalid coins %v", txOut.Coins))
	}
	if txOut.Coins.IsZero() {
		return abci.ErrBaseInvalidOutput.AppendLog("Coins cannot be zero")
	}
	return abci.OK
}

func (txOut TxOutput) String() string {
	return Fmt("TxOutput{%X,%v}", txOut.Address, txOut.Coins)
}

//-----------------------------------------------------------------------------

type SendTx struct {
	Gas     int64      `json:"gas"` // Gas
	Fee     Coin       `json:"fee"` // Fee
	Inputs  []TxInput  `json:"inputs"`
	Outputs []TxOutput `json:"outputs"`
}

func (tx *SendTx) SignBytes(chainID string) []byte {
	signBytes := wire.BinaryBytes(chainID)
	sigz := make([]crypto.Signature, len(tx.Inputs))
	for i := range tx.Inputs {
		sigz[i] = tx.Inputs[i].Signature
		tx.Inputs[i].Signature = crypto.Signature{}
	}
	signBytes = append(signBytes, wire.BinaryBytes(tx)...)
	for i := range tx.Inputs {
		tx.Inputs[i].Signature = sigz[i]
	}
	return signBytes
}

func (tx *SendTx) SetSignature(addr []byte, sig crypto.Signature) bool {
	for i, input := range tx.Inputs {
		if bytes.Equal(input.Address, addr) {
			tx.Inputs[i].Signature = sig
			return true
		}
	}
	return false
}

func (tx *SendTx) String() string {
	return Fmt("SendTx{%v/%v %v->%v}", tx.Gas, tx.Fee, tx.Inputs, tx.Outputs)
}

//-----------------------------------------------------------------------------

type AppTx struct {
	Gas   int64           `json:"gas"`   // Gas
	Fee   Coin            `json:"fee"`   // Fee
	Name  string          `json:"type"`  // Which plugin
	Input TxInput         `json:"input"` // Hmmm do we want coins?
	Data  json.RawMessage `json:"data"`
}

func (tx *AppTx) SignBytes(chainID string) []byte {
	signBytes := wire.BinaryBytes(chainID)
	sig := tx.Input.Signature
	tx.Input.Signature = crypto.Signature{}
	signBytes = append(signBytes, wire.BinaryBytes(tx)...)
	tx.Input.Signature = sig
	return signBytes
}

func (tx *AppTx) SetSignature(sig crypto.Signature) bool {
	tx.Input.Signature = sig
	return true
}

func (tx *AppTx) String() string {
	return Fmt("AppTx{%v/%v %v %v %X}", tx.Gas, tx.Fee, tx.Name, tx.Input, tx.Data)
}

//-----------------------------------------------------------------------------

type PostTx struct {
	Address   data.Bytes       `json:"address"`   // Hash of the PubKey
	Title     string           `json:"title"`
	Content   string           `json:"content"`
	Sequence  int              `json:"sequence"`  // Must be 1 greater than the last committed PostTx
	Parent    []byte           `json:"parent"`
	Signature crypto.Signature `json:"signature"` // Depends on the PubKey type and the whole Tx
	PubKey    crypto.PubKey    `json:"pub_key"`   // Is present iff Sequence == 0
}

func (tx *PostTx) SignBytes(chainID string) []byte {
	signBytes := wire.BinaryBytes(chainID)
	sig := tx.Signature
	tx.Signature = crypto.Signature{}
	signBytes = append(signBytes, wire.BinaryBytes(tx)...)
	tx.Signature = sig
	return signBytes
}

func (tx *PostTx) SetSignature(sig crypto.Signature) bool {
	tx.Signature = sig
	return true
}

func (tx PostTx) ValidateBasic() abci.Result {
	if len(tx.Address) != 20 {
		return abci.ErrBaseInvalidInput.AppendLog("Invalid address length")
	}
	if tx.Sequence <= 0 {
		return abci.ErrBaseInvalidInput.AppendLog("Sequence must be greater than 0")
	}
	if tx.Sequence == 1 && tx.PubKey.Empty() {
		return abci.ErrBaseInvalidInput.AppendLog("PubKey must be present when Sequence == 1")
	}
	if tx.Sequence > 1 && !tx.PubKey.Empty() {
		return abci.ErrBaseInvalidInput.AppendLog("PubKey must be nil when Sequence > 1")
	}
	return abci.OK
}

func (tx *PostTx) String() string {
	return Fmt("PostTx{%v, %v, %v, %v}", tx.Address, tx.Title, tx.Content, tx.Sequence)
}

//-----------------------------------------------------------------------------

type DonateTx struct {
	Input     TxInput          `json:"inputTx"` // Hmmm do we want coins?
    To        []byte           `json:"to"`      //post_id
    Fee       Coin             `json:"fee"`
}

func (tx *DonateTx) SignBytes(chainID string) []byte {
	signBytes := wire.BinaryBytes(chainID)
	sig := tx.Input.Signature
	tx.Input.Signature = crypto.Signature{}
	signBytes = append(signBytes, wire.BinaryBytes(tx)...)
	tx.Input.Signature = sig
	return signBytes
}

func (tx *DonateTx) SetSignature(sig crypto.Signature) bool {
    tx.Input.Signature = sig
    return true
}

func (tx DonateTx) ValidateBasic() abci.Result {
	if err := tx.Input.ValidateBasic(); !err.IsErr() {
		return err
	}
	return abci.OK
}

func (tx *DonateTx) String() string {
        return Fmt("DonateTx{%v -> %v, %v}", tx.Input, tx.To, tx.Fee)
}

//-----------------------------------------------------------------------------
// feature-like
type LikeTx struct {
	From      data.Bytes       `json:"from"`      // address
	To        []byte           `json:"to"`        // post_id
	Weight    int              `json:"weight"`    // like weight from -10000 to 10000
	Signature crypto.Signature `json:"signature"` // Depends on the PubKey type and the whole Tx
	PubKey    crypto.PubKey    `json:"pub_key"`   // Is present iff Sequence == 0
}

func (tx *LikeTx) SignBytes(chainID string) []byte {
	signBytes := wire.BinaryBytes(chainID)
	sig := tx.Signature
	tx.Signature = crypto.Signature{}
	signBytes = append(signBytes, wire.BinaryBytes(tx)...)
	tx.Signature = sig
	return signBytes
}

func (tx *LikeTx) SetSignature(sig crypto.Signature) bool {
	tx.Signature = sig
	return true
}

func (tx LikeTx) ValidateBasic() abci.Result {
	if len(tx.From) != 20 {
		return abci.ErrBaseInvalidInput.AppendLog("Invalid address length")
	}
	if tx.Weight < -10000 || tx.Weight > 10000 {
		return abci.ErrBaseInvalidInput.AppendLog("Invalid weight")
	}
	return abci.OK
}

func (tx *LikeTx) String() string {
	return Fmt("LikeTx{ set %v from %v to %v. (%v, %v) }",
		tx.Weight, tx.From, tx.To, tx.Signature, tx.PubKey)
}

func TxID(chainID string, tx Tx) []byte {
	signBytes := tx.SignBytes(chainID)
	return wire.BinaryRipemd160(signBytes)
}

//--------------------------------------------------------------------------------

// Contract: This function is deterministic and completely reversible.
func jsonEscape(str string) string {
	escapedBytes, err := json.Marshal(str)
	if err != nil {
		PanicSanity(Fmt("Error json-escaping a string", str))
	}
	return string(escapedBytes)
}