package shared

var InvoiceTypes = []InvoiceType{
	WriteOff,
	AddCredit,
	AddDebit,
	Unapply,
	Reapply,
}

type InvoiceType int

const (
	Unknown InvoiceType = iota
	WriteOff
	AddCredit
	AddDebit
	Unapply
	Reapply
)

func (i InvoiceType) Translation() string {
	switch i {
	case WriteOff:
		return "Write off"
	case AddCredit:
		return "Add credit"
	case AddDebit:
		return "Add debit"
	case Unapply:
		return "Unapply"
	case Reapply:
		return "Reapply"
	default:
		return ""
	}
}

func (i InvoiceType) Key() string {
	switch i {
	case WriteOff:
		return "CREDIT WRITE OFF" // ?
	case AddCredit:
		return "CREDIT MEMO" // ?
	case AddDebit:
		return "UNKNOWN DEBIT" // ?
	case Unapply:
		return "UNAPPLY" // ?
	case Reapply:
		return "REAPPLY" // ?
	default:
		return ""
	}
}

func (i InvoiceType) AmountRequired() bool {
	switch i {
	case AddCredit, AddDebit, Unapply, Reapply:
		return true
	default:
		return false
	}
}
