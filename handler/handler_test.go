package handler_test

import (
	"reflect"
	"testing"
	"time"

	openapi_types "github.com/discord-gophers/goapi-gen/types"
	"github.com/joshinjohnson/wave-exercise/handler"
	"github.com/joshinjohnson/wave-exercise/pkg/payroll"
)

func TestConvertReport(t *testing.T) {
	mockReport := payroll.PayrollReport{
		EmployeeReports: []payroll.EmployeeReport{
			{
				AmountPaid: 100.0,
				EmployeeId: 1,
				PayPeriod:  payroll.PayPeriod{StartDate: time.Now(), EndDate: time.Now().AddDate(0, 0, 14)},
			},
		},
	}

	expectedConvertedReport := handler.PayrollReport{
		EmployeeReports: []handler.WorkerPayrollBiWeek{
			{
				AmountPaid: "$100.00",
				EmployeeID: 1,
				PayPeriod: struct {
					EndDate   *openapi_types.Date `json:"end_date,omitempty"`
					StartDate *openapi_types.Date `json:"start_date,omitempty"`
				}{
					StartDate: handler.ConvertDate(mockReport.EmployeeReports[0].PayPeriod.StartDate),
					EndDate:   handler.ConvertDate(mockReport.EmployeeReports[0].PayPeriod.EndDate),
				},
			},
		},
	}

	actualConvertedReport := handler.ConvertReport(mockReport)

	if !reflect.DeepEqual(actualConvertedReport, expectedConvertedReport) {
		t.Errorf("Conversion result mismatch. Expected:\n%v\n Got:\n%v", expectedConvertedReport, actualConvertedReport)
	}
}

func TestConvertDate(t *testing.T) {
	mockTime := time.Date(2022, time.January, 1, 0, 0, 0, 0, time.Local)

	expectedConvertedDate := &openapi_types.Date{
		Time: mockTime,
	}

	actualConvertedDate := handler.ConvertDate(mockTime)

	if !actualConvertedDate.Equal(expectedConvertedDate.Time) {
		t.Errorf("Date conversion result mismatch. Expected:\n%v\n Got:\n%v", expectedConvertedDate, actualConvertedDate)
	}
}

func TestParseTime(t *testing.T) {
	tests := []struct {
		input    string
		expected *time.Time
		errMsg   string
	}{
		{"01/02/2023", createTime(2023, 2, 1, 0, 0, 0), ""},
		{"31/12/2022", createTime(2022, 12, 31, 0, 0, 0), ""},
		{"02/29/2021", nil, "invalid date specified"},
		{"01/01/twothousand", nil, "invalid date specified"},
		{"", nil, "invalid date specified"},
		{"01/01", nil, "invalid date specified"},
		{"01/01/2022/2022", nil, "invalid date specified"},
		{"01-01-2022", nil, "invalid date specified"},
		{"01//2022", nil, "invalid date specified"},
		{"01/13/9999", nil, "invalid date specified"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := handler.ParseTime(test.input)

			if err != nil && test.errMsg != err.Error()[:len(test.errMsg)] {
				t.Errorf("Expected error message: %s, but got: %s", test.errMsg, err.Error())
			}

			if !timesEqual(result, test.expected) {
				t.Errorf("Expected time: %v, but got: %v", test.expected, result)
			}
		})
	}
}

func timesEqual(t1, t2 *time.Time) bool {
	if t1 == nil && t2 == nil {
		return true
	}
	if t1 == nil || t2 == nil {
		return false
	}
	return t1.Equal(*t2)
}

func createTime(year, month, day, hour, min, sec int) *time.Time {
	t := time.Date(year, time.Month(month), day, hour, min, sec, 0, time.Local)
	return &t
}
