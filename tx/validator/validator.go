package validator

import (
	acc "github.com/lino-network/lino/tx/account"
	types "github.com/lino-network/lino/types"
	abci "github.com/tendermint/abci/types"
)

// Validator is basic structure records all validator information
type Validator struct {
	ABCIValidator       abci.Validator
	Username            acc.AccountKey `json:"username"`
	Deposit             types.Coin     `json:"deposit"`
	AbsentVote          int            `json:"absent_vote"`
	WithdrawAvailableAt types.Height   `json:"withdraw_available_at"`
	IsByzantine         bool           `json:"is_byzantine"`
}

// Validator list
type ValidatorList struct {
	OncallValidators []acc.AccountKey `json:"oncall_validators"`
	AllValidators    []acc.AccountKey `json:"all_validators"`
	LowestPower      types.Coin       `json:"lowest_power"`
	LowestValidator  acc.AccountKey   `json:"lowest_validator"`
}

var valRegisterFee = types.Coin{Amount: 1000 * types.Decimals}
