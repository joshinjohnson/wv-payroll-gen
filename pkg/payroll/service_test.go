package payroll_test

import (
	"testing"
	"time"

	"github.com/joshinjohnson/wave-exercise/pkg/payroll"
	"github.com/stretchr/testify/assert"
)

func TestGenerateReport(t *testing.T) {
	jobGroupRates := []payroll.JobGroupRate{
		{JobGroup: "A", Rate: 15.0},
		{JobGroup: "B", Rate: 20.0},
	}
	worklogs := []payroll.WorkLog{
		{EmployeeId: 1, Date: time.Date(2023, 1, 2, 0, 0, 0, 0, time.Local), HoursLogged: 8, JobGroup: "A"},
		{EmployeeId: 1, Date: time.Date(2023, 1, 18, 0, 0, 0, 0, time.Local), HoursLogged: 4, JobGroup: "B"},
	}

	report := payroll.GenerateReport(jobGroupRates, worklogs)

	assert.NotNil(t, report)
	assert.Equal(t, len(worklogs), len(report.EmployeeReports))
}

func TestParsePayPeriodString(t *testing.T) {
	payPeriodString := "1-1-2023"
	payPeriod := payroll.ParsePayPeriodString(payPeriodString)

	assert.NotNil(t, payPeriod)
	assert.Equal(t, time.Date(2023, 1, 1, 0, 0, 0, 0, time.Local), payPeriod.StartDate)
	assert.Equal(t, time.Date(2023, 1, 15, 0, 0, 0, 0, time.Local), payPeriod.EndDate)
}

func TestCalcAmountPaid(t *testing.T) {
	groupRates := map[payroll.JobGroup]float64{
		payroll.GroupA: 15.0,
		payroll.GroupB: 20.0,
	}

	logs := []payroll.WorkLog{
		{HoursLogged: 8, JobGroup: "A"},
		{HoursLogged: 4, JobGroup: "B"},
	}

	amountPaid := payroll.CalcAmountPaid(groupRates, logs)

	assert.Equal(t, 8*15.0+4*20.0, amountPaid)
}

func TestGetPayPeriodString(t *testing.T) {
	date1 := time.Date(2023, 1, 2, 0, 0, 0, 0, time.Local)
	date2 := time.Date(2023, 1, 18, 0, 0, 0, 0, time.Local)

	payPeriodString1 := payroll.GetPayPeriodString(date1)
	payPeriodString2 := payroll.GetPayPeriodString(date2)

	assert.Equal(t, "1-1-2023", payPeriodString1)
	assert.Equal(t, "16-1-2023", payPeriodString2)
}

func TestDaysInMonth(t *testing.T) {
	type args struct {
		year  int
		month time.Month
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "positive",
			args: args{
				year:  2023,
				month: time.November,
			},
			want: 30,
		},
		{
			name: "positive",
			args: args{
				year:  2023,
				month: time.February,
			},
			want: 28,
		},
		{
			name: "positive",
			args: args{
				year:  2023,
				month: time.December,
			},
			want: 31,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := payroll.DaysInMonth(tt.args.year, tt.args.month); got != tt.want {
				t.Errorf("DaysInMonth() = %v, want %v", got, tt.want)
			}
		})
	}
}
