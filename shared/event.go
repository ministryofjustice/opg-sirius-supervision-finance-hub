package shared

import (
	"encoding/json"
	"fmt"
)

const (
	EventSourceSirius                   = "opg.supervision.sirius"
	EventSourceFinanceAdhoc             = "opg.supervision.finance.adhoc"
	EventSourceInfra                    = "opg.supervision.infra"
	DetailTypeFinanceAdhoc              = "finance-adhoc"
	DetailTypeDebtPositionChanged       = "debt-position-changed"
	DetailTypeClientCreated             = "client-created"
	DetailTypeClientMadeInactive        = "client-made-inactive"
	DetailTypeFinanceAdminUpload        = "finance-admin-upload"
	DetailTypeScheduledEvent            = "scheduled-event"
	ScheduledEventRefundExpiry          = "refund-expiry"
	ScheduledEventDirectDebitCollection = "direct-debit-collection"
	ScheduledEventFailedCollections     = "failed-direct-debit-collections"
)

type Event struct {
	Source       string      `json:"source"`
	EventBusName string      `json:"event-bus-name"`
	DetailType   string      `json:"detail-type"`
	Detail       interface{} `json:"detail"`
}

func (e *Event) UnmarshalJSON(data []byte) error {
	type tmp Event // avoids infinite recursion
	if err := json.Unmarshal(data, (*tmp)(e)); err != nil {
		return err
	}

	var raw struct {
		Detail json.RawMessage `json:"detail"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	switch e.DetailType {
	case DetailTypeDebtPositionChanged:
		var detail DebtPositionChangedEvent
		if err := json.Unmarshal(raw.Detail, &detail); err != nil {
			return err
		}
		e.Detail = detail
	case DetailTypeFinanceAdminUpload:
		var detail FinanceAdminUploadEvent
		if err := json.Unmarshal(raw.Detail, &detail); err != nil {
			return err
		}
		e.Detail = detail
	case DetailTypeClientCreated:
		var detail ClientCreatedEvent
		if err := json.Unmarshal(raw.Detail, &detail); err != nil {
			return err
		}
		e.Detail = detail
	case DetailTypeClientMadeInactive:
		var detail ClientMadeInactiveEvent
		if err := json.Unmarshal(raw.Detail, &detail); err != nil {
			return err
		}
		e.Detail = detail
	case DetailTypeFinanceAdhoc:
		var detail AdhocEvent
		if err := json.Unmarshal(raw.Detail, &detail); err != nil {
			return err
		}
		e.Detail = detail
	case DetailTypeScheduledEvent:
		var detail ScheduledEvent
		if err := json.Unmarshal(raw.Detail, &detail); err != nil {
			return err
		}
		e.Detail = detail
	default:
		return fmt.Errorf("unknown detail type: %s", e.DetailType)
	}

	return nil
}

type DebtPositionChangedEvent struct {
	ClientID int32 `json:"clientId"`
}

type ClientCreatedEvent struct {
	ClientID int32  `json:"clientId"`
	CourtRef string `json:"courtRef"`
}

type ClientMadeInactiveEvent struct {
	ClientID int32  `json:"clientId"`
	CourtRef string `json:"courtRef"`
	Surname  string `json:"surname"`
}

type FinanceAdminUploadEvent struct {
	EmailAddress string           `json:"emailAddress"`
	Filename     string           `json:"filename"`
	UploadType   ReportUploadType `json:"uploadType"`
	UploadDate   Date             `json:"uploadDate"`
	PisNumber    int              `json:"pisNumber"`
}

type AdhocEvent struct {
	Task string `json:"task"`
}

type ScheduledEvent struct {
	Trigger  string      `json:"trigger"`
	Override interface{} `json:"override"`
}

func (e *ScheduledEvent) UnmarshalJSON(data []byte) error {
	type tmp ScheduledEvent // avoids infinite recursion
	if err := json.Unmarshal(data, (*tmp)(e)); err != nil {
		return err
	}

	var raw struct {
		Override json.RawMessage `json:"override"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	switch e.Trigger {
	case ScheduledEventDirectDebitCollection, ScheduledEventFailedCollections:
		var override DateOverride
		if err := json.Unmarshal(raw.Override, &override); err != nil {
			return err
		}
		e.Override = override
	case ScheduledEventRefundExpiry:
		e.Override = nil
	default:
		return fmt.Errorf("unknown override type: %s", e.Override)
	}

	return nil
}

type DateOverride struct {
	Date Date `json:"date"`
}
