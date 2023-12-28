package payroll

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/joshinjohnson/wave-exercise/pkg/db"
	"github.com/sirupsen/logrus"
)

type payrollService struct {
	payrollRepo *payrollRepository
}

func NewPayrollService(dbW *db.DbWrapper) payrollService {
	return payrollService{
		payrollRepo: NewPayrollRepository(dbW),
	}
}

func (s payrollService) GetReport(limit, offset uint64) (PayrollReport, error) {
	groupRates, err := s.payrollRepo.GetJobGroupRates()
	if err != nil {
		return PayrollReport{}, ErrReportGenerate
	}

	worklogs, err := s.payrollRepo.Get(limit, offset)
	if err != nil {
		return PayrollReport{}, ErrReportGenerate
	}

	return GenerateReport(groupRates, worklogs), nil
}

func (s payrollService) InsertLogs(filenameId int, logs []WorkLog) error {
	tx, err := s.payrollRepo.dbW.DB.Begin()
	if err != nil {
		return fmt.Errorf("error while starting tx: %v", err)
	}
	s.payrollRepo.dbW.Tx = tx

	if err := s.payrollRepo.InsertFileId(filenameId); err != nil {
		return ErrFileIdExists
	}

	ids, err := s.payrollRepo.CreateN(logs)
	if err != nil {
		logrus.Errorf("error while inserting logs: %v", err)
		return ErrWorkLogCreate
	}

	logrus.Infof(fmt.Sprintf("created log ids: %d", ids))
	return nil
}

func GenerateReport(jobGroupRates []JobGroupRate, worklogs []WorkLog) PayrollReport {
	empPerPeriodData := make(map[int]map[string][]WorkLog)
	for _, worklog := range worklogs {
		if _, ok := empPerPeriodData[worklog.EmployeeId]; !ok {
			empPerPeriodData[worklog.EmployeeId] = make(map[string][]WorkLog)
		}

		empPerPeriodData[worklog.EmployeeId][GetPayPeriodString(worklog.Date)] =
			append(empPerPeriodData[worklog.EmployeeId][GetPayPeriodString(worklog.Date)], worklog)
	}

	groupRatesMap := make(map[JobGroup]float64)
	for _, jobGroupRate := range jobGroupRates {
		groupRatesMap[jobGroupRate.JobGroup] = jobGroupRate.Rate
	}

	empReports := make([]EmployeeReport, 0, len(worklogs))
	for empId, payPeriodData := range empPerPeriodData {
		for payPeriod, workLogs := range payPeriodData {
			empReports = append(empReports, EmployeeReport{
				EmployeeId: empId,
				PayPeriod:  ParsePayPeriodString(payPeriod),
				AmountPaid: CalcAmountPaid(groupRatesMap, workLogs),
			})
		}
	}

	return PayrollReport{
		EmployeeReports: empReports,
	}
}

func ParsePayPeriodString(logs string) PayPeriod {
	parts := strings.Split(logs, "-")
	dayStart, _ := strconv.Atoi(parts[0])
	month, _ := strconv.Atoi(parts[1])
	year, _ := strconv.Atoi(parts[2])
	monthEnum := GetMonth(month)

	dayEnd := DaysInMonth(year, monthEnum)
	if dayStart == 1 {
		dayEnd = 15
	}

	return PayPeriod{
		StartDate: time.Date(year, monthEnum, dayStart, 0, 0, 0, 0, time.Local),
		EndDate:   time.Date(year, monthEnum, dayEnd, 0, 0, 0, 0, time.Local),
	}
}

// daysInMonth func checks days in a month, by setting day to 0 and month to next month,
// so it normalizes to prev month last day
func DaysInMonth(year int, month time.Month) int {
	lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC)
	return lastDay.Day()
}

func CalcAmountPaid(groupRatesMap map[JobGroup]float64, logs []WorkLog) float64 {
	var totalAmountPaid float64

	for _, log := range logs {
		totalAmountPaid += float64(log.HoursLogged) * groupRatesMap[log.JobGroup]
	}

	return totalAmountPaid
}

// generates startdate-month-year string for a time object
func GetPayPeriodString(date time.Time) string {
	if date.Day() >= 1 && date.Day() <= 15 {
		return fmt.Sprintf("1-%v-%v", int(date.Month()), date.Year())
	}
	return fmt.Sprintf("16-%v-%v", int(date.Month()), date.Year())
}

func GetMonth(month int) time.Month {
	switch month {
	case 1:
		return time.January
	case 2:
		return time.February
	case 3:
		return time.March
	case 4:
		return time.April
	case 5:
		return time.May
	case 6:
		return time.June
	case 7:
		return time.July
	case 8:
		return time.August
	case 9:
		return time.September
	case 10:
		return time.October
	case 11:
		return time.November
	case 12:
		return time.December
	}

	return time.January
}

// Requirements:
// 1. job group A is paid $20/hr, and job group B is paid $30/hr
// 2.  If an attempt is made to upload a file with the same report ID as a previously uploaded file, this upload should fail with an error message indicating that this is not allowed.
// 3. The payPeriod field is an object containing a date interval that is roughly biweekly. Each month has two pay periods; the first half is from the 1st to the 15th inclusive, and the second half is from the 16th to the end of the month, inclusive. payPeriod will have two fields to represent this interval: startDate and endDate.
// 4. Unique rows: "employeeId": "1", "payPeriod": { "startDate": "2023-01-01", "endDate": "2023-01-15"}
// 4. The report should be sorted in some sensical order (e.g. sorted by employee id and then pay period start.)
