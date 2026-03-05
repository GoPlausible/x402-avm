package facilitator

// Facilitator error constants for the exact AVM (Algorand) scheme (V2)
const (
	// Verify errors
	ErrUnsupportedScheme              = "invalid_exact_algorand_unsupported_scheme"
	ErrNetworkMismatch                = "invalid_exact_algorand_network_mismatch"
	ErrMissingFeePayer                = "invalid_exact_algorand_payload_missing_fee_payer"
	ErrFeePayerNotManaged             = "invalid_exact_algorand_fee_payer_not_managed_by_facilitator"
	ErrInvalidPayloadFormat           = "invalid_exact_algorand_payload_format"
	ErrInvalidPaymentIndex            = "invalid_exact_algorand_payload_invalid_payment_index"
	ErrGroupSizeExceeded              = "invalid_exact_algorand_payload_group_size_exceeded"
	ErrInvalidTransaction             = "invalid_exact_algorand_payload_invalid_transaction"
	ErrInvalidGroupID                 = "invalid_exact_algorand_payload_invalid_group_id"
	ErrPaymentNotAssetTransfer        = "invalid_exact_algorand_payload_payment_not_asset_transfer"
	ErrAmountMismatch                 = "invalid_exact_algorand_payload_amount_mismatch"
	ErrReceiverMismatch               = "invalid_exact_algorand_payload_receiver_mismatch"
	ErrAssetMismatch                  = "invalid_exact_algorand_payload_asset_mismatch"
	ErrInvalidFeePayer                = "invalid_exact_algorand_invalid_fee_payer"
	ErrFeePayerInvalidType            = "invalid_exact_algorand_fee_payer_invalid_type"
	ErrFeePayerNonZeroAmount          = "invalid_exact_algorand_fee_payer_non_zero_amount"
	ErrFeePayerNotSelfPayment         = "invalid_exact_algorand_fee_payer_not_self_payment"
	ErrFeePayerHasCloseTo             = "invalid_exact_algorand_fee_payer_has_close_to"
	ErrFeePayerHasRekeyTo             = "invalid_exact_algorand_fee_payer_has_rekey_to"
	ErrFeeTooHigh                     = "invalid_exact_algorand_fee_too_high"
	ErrPaymentNotSigned               = "invalid_exact_algorand_payload_payment_not_signed"
	ErrInvalidSignature               = "invalid_exact_algorand_payload_invalid_signature"
	ErrSimulationFailed               = "invalid_exact_algorand_simulation_failed"
	ErrFacilitatorTransferringFunds   = "invalid_exact_algorand_facilitator_transferring_funds"
	ErrGenesisHashMismatch            = "invalid_exact_algorand_genesis_hash_mismatch"
	ErrUnsignedNonFacilitatorTxn      = "invalid_exact_algorand_unsigned_non_facilitator_txn"
	ErrSecurityRekeyNotAllowed        = "invalid_exact_algorand_security_rekey_not_allowed"
	ErrSecurityCloseToNotAllowed      = "invalid_exact_algorand_security_close_to_not_allowed"
	ErrSecurityKeyregNotAllowed       = "invalid_exact_algorand_security_keyreg_not_allowed"

	// Settle errors
	ErrVerificationFailed             = "invalid_exact_algorand_verification_failed"
	ErrTransactionSigningFailed       = "invalid_exact_algorand_transaction_signing_failed"
	ErrTransactionFailed              = "invalid_exact_algorand_transaction_failed"
	ErrTransactionConfirmationFailed  = "invalid_exact_algorand_transaction_confirmation_failed"
)
