package developer

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lino-network/lino/types"
)

// Error constructors
func ErrDeveloperNotFound() sdk.Error {
	return sdk.NewError(types.CodeUsernameNotFound, fmt.Sprintf("Developer not found"))
}

func ErrDeveloperExist(username types.AccountKey) sdk.Error {
	return sdk.NewError(types.CodeDeveloperHandlerFailed, fmt.Sprintf("Developer %v exist", username))
}

func ErrUsernameNotFound() sdk.Error {
	return sdk.NewError(types.CodeUsernameNotFound, fmt.Sprintf("Username not found"))
}

func ErrDeveloperDepositNotEnough() sdk.Error {
	return sdk.NewError(types.CodeDeveloperManagerFailed, fmt.Sprintf("Developer deposit not enough"))
}

func ErrInvalidUsername() sdk.Error {
	return sdk.NewError(types.CodeInvalidUsername, fmt.Sprintf("Invalida Username"))
}

func ErrNoCoinToWithdraw() sdk.Error {
	return sdk.NewError(types.CodeDeveloperManagerFailed, fmt.Sprintf("No coin to withdraw"))
}

func ErrInvalidValidityPeriod() sdk.Error {
	return sdk.NewError(types.CodeInvalidMsg, fmt.Sprintf("invalid grant validity period"))
}