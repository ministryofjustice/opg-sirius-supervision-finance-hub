package shared

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReportUploadType_Filename(t *testing.T) {
	tests := []struct {
		name         string
		uploadType   ReportUploadType
		dateString   string
		wantErr      bool
		wantFilename string
	}{
		{
			name:         "Non-moto card payments report type",
			uploadType:   ReportTypeUploadDeputySchedule,
			dateString:   "2020-01-02",
			wantErr:      false,
			wantFilename: "",
		},
		{
			name:         "Moto card payments report type",
			uploadType:   ReportTypeUploadPaymentsMOTOCard,
			dateString:   "2020-01-02",
			wantErr:      false,
			wantFilename: "feemoto_02:01:2020normal.csv",
		},
		{
			name:         "Online card payments report type",
			uploadType:   ReportTypeUploadPaymentsOnlineCard,
			dateString:   "2024-12-03",
			wantErr:      false,
			wantFilename: "feemoto_03:12:2024mlpayments.csv",
		},
		{
			name:         "Supervision BACS payments report type",
			uploadType:   ReportTypeUploadPaymentsSupervisionBACS,
			dateString:   "2024-10-01",
			wantErr:      false,
			wantFilename: "feebacs_01:10:2024_new_acc.csv",
		},
		{
			name:         "OPG BACS payments report type",
			uploadType:   ReportTypeUploadPaymentsOPGBACS,
			dateString:   "2024-11-21",
			wantErr:      false,
			wantFilename: "feebacs_21:11:2024.csv",
		},
		{
			name:         "Invalid date",
			uploadType:   ReportTypeUploadPaymentsMOTOCard,
			dateString:   "02/01/2020",
			wantErr:      true,
			wantFilename: "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			filename, err := test.uploadType.Filename(test.dateString)

			assert.Equal(t, test.wantFilename, filename)

			if test.wantErr {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
