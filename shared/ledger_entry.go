package shared

type CreateLedgerEntryRequest struct {
	AdjustmentType AdjustmentType `json:"adjustmentType" validate:"valid-enum"`
	Notes          string         `json:"notes" validate:"thousand-character-limit"`
	Amount         int            `json:"amount" validate:"gt=0"`
}
