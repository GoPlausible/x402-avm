package server

// Server error constants for the exact AVM (Algorand) scheme (V2)
const (
	ErrAmountMustBeString    = "invalid_exact_algorand_server_amount_must_be_string"
	ErrFailedToParsePrice    = "invalid_exact_algorand_server_failed_to_parse_price"
	ErrInvalidPriceFormat    = "invalid_exact_algorand_server_invalid_price_format"
	ErrFailedToConvertAmount = "invalid_exact_algorand_server_failed_to_convert_amount"
	ErrFailedToParseAmount   = "invalid_exact_algorand_server_failed_to_parse_amount"
)
