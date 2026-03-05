package client

// Client error constants for the exact AVM (Algorand) scheme (V1)
const (
	ErrUnsupportedNetwork             = "invalid_exact_algorand_v1_client_unsupported_network"
	ErrInvalidAssetID                 = "invalid_exact_algorand_v1_client_invalid_asset_id"
	ErrInvalidPayToAddress            = "invalid_exact_algorand_v1_client_invalid_payto_address"
	ErrInvalidAmount                  = "invalid_exact_algorand_v1_client_invalid_amount"
	ErrFeePayerRequired               = "invalid_exact_algorand_v1_client_fee_payer_required"
	ErrInvalidFeePayerAddress         = "invalid_exact_algorand_v1_client_invalid_fee_payer_address"
	ErrInvalidExtraField              = "invalid_exact_algorand_v1_client_invalid_extra_field"
	ErrFailedToGetTransactionParams   = "invalid_exact_algorand_v1_client_failed_to_get_transaction_params"
	ErrFailedToBuildFeePayerTxn       = "invalid_exact_algorand_v1_client_failed_to_build_fee_payer_txn"
	ErrFailedToBuildAssetTransferTxn  = "invalid_exact_algorand_v1_client_failed_to_build_asset_transfer_txn"
	ErrFailedToAssignGroupID          = "invalid_exact_algorand_v1_client_failed_to_assign_group_id"
	ErrFailedToSignTransaction        = "invalid_exact_algorand_v1_client_failed_to_sign_transaction"
	ErrFailedToEncodeTransaction      = "invalid_exact_algorand_v1_client_failed_to_encode_transaction"
)
