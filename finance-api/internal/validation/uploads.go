package validation

const (
	UploadErrorDateParse              = "DATE_PARSE_ERROR"
	UploadErrorDateTimeParse          = "DATE_TIME_PARSE_ERROR"
	UploadErrorAmountParse            = "AMOUNT_PARSE_ERROR"
	UploadErrorClientNotFound         = "CLIENT_NOT_FOUND"
	UploadErrorPaymentTypeParse       = "PAYMENT_TYPE_PARSE_ERROR"
	UploadErrorUnknownUploadType      = "UNKNOWN_UPLOAD_TYPE"
	UploadErrorNoMatchedPayment       = "NO_MATCHED_PAYMENT"
	UploadErrorReversalClientNotFound = "REVERSAL_CLIENT_NOT_FOUND"
	UploadErrorDuplicateReversal      = "DUPLICATE_REVERSAL"
)
