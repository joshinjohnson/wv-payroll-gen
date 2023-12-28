package payroll

import "time"

type JobGroup string

const (
	GroupA JobGroup = "A"
	GroupB JobGroup = "B"
)

type WorkLog struct {
	EmployeeId  int
	JobGroup    JobGroup
	Date        time.Time
	HoursLogged float64
}

type PayrollReport struct {
	EmployeeReports []EmployeeReport
}

type EmployeeReport struct {
	EmployeeId int
	PayPeriod  PayPeriod
	AmountPaid float64
}

type PayPeriod struct {
	StartDate time.Time
	EndDate   time.Time
}

type JobGroupRate struct {
	JobGroup JobGroup
	Rate     float64
}
