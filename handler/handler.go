package handler

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/discord-gophers/goapi-gen/types"
	openapi_types "github.com/discord-gophers/goapi-gen/types"
	"github.com/joshinjohnson/wave-exercise/pkg/payroll"
	"github.com/sirupsen/logrus"
)

type PayrollHandler struct {
	payrollService PayrollService
}

func NewPayrollHandler(payrollService PayrollService) PayrollHandler {
	return PayrollHandler{
		payrollService: payrollService,
	}
}

func (h PayrollHandler) GetReport(http.ResponseWriter, *http.Request) *Response {
	report, err := h.payrollService.GetReport(1000, 0)
	if err != nil {
		logrus.Errorf("error while generating report: %v", err)
		return GetReportJSON500Response(Error{})
	}

	return GetReportJSON200Response(ConvertReport(report))
}

func (h PayrollHandler) PostUpload(w http.ResponseWriter, r *http.Request) *Response {
	// Parse the multipart form data
	// 10 MB maximum file size
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		logrus.Errorf("error while parsing csv: %v", err)
		return PostUploadJSON400Response(Error{
			Message: ErrCSVFileProcessingError,
		})
	}

	// Get the file from the request
	file, handler, err := r.FormFile("file")
	if err != nil {
		logrus.Errorf("error while parsing csv: %v", err)
		return PostUploadJSON400Response(Error{
			Message: ErrCSVFileProcessingError,
		})
	}
	defer file.Close()

	// get csv file id
	filename := strings.ReplaceAll(handler.Filename, ".csv", "")
	filenameParts := strings.Split(filename, "-")
	if len(filenameParts) < 3 {
		logrus.Errorf("error while parsing file name")
		return PostUploadJSON400Response(Error{
			Message: ErrCSVFileProcessingError,
		})
	}

	filenameId, err := strconv.Atoi(filenameParts[2])
	if err != nil {
		logrus.Errorf("error reading CSV file: %v", err)
		return PostUploadJSON400Response(Error{
			Message: ErrCSVFileProcessingError,
		})
	}

	// Create a CSV reader
	reader := csv.NewReader(file)

	// Ignore header
	reader.Read()

	serviceWorkLogs := make([]payroll.WorkLog, 0)
	// Process each row line by line
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			logrus.Errorf("error reading CSV file: %v", err)
			return PostUploadJSON400Response(Error{
				Message: ErrCSVFileProcessingError,
			})
		}

		if len(row) < 3 {
			logrus.Errorf("error reading CSV file: no. of rows is less than 3")
			return PostUploadJSON400Response(Error{
				Message: ErrCSVFileProcessingError,
			})
		}

		logDate, err := ParseTime(row[0])
		if err != nil || logDate == nil {
			logrus.Errorf("error reading CSV file: %v", err)
			return PostUploadJSON400Response(Error{
				Message: ErrCSVFileProcessingError,
			})
		}

		logHours, err := strconv.ParseFloat(row[1], 64)
		if err != nil {
			logrus.Errorf("error reading CSV file: %v", err)
			return PostUploadJSON400Response(Error{
				Message: ErrCSVFileProcessingError,
			})
		}

		employeeID, err := strconv.ParseUint(row[2], 10, 64)
		if err != nil {
			logrus.Errorf("error reading CSV file: %v", err)
			return PostUploadJSON400Response(Error{
				Message: ErrCSVFileProcessingError,
			})
		}

		logJobGroup := ConvertWorkGroup(row[3])

		serviceWorkLogs = append(serviceWorkLogs, payroll.WorkLog{
			EmployeeId:  int(employeeID),
			HoursLogged: logHours,
			JobGroup:    logJobGroup,
			Date:        *logDate,
		})
	}

	err = h.payrollService.InsertLogs(filenameId, serviceWorkLogs)
	if err != nil {
		logrus.Errorf("error while inserting logs: %v", err)
		return PostUploadJSON500Response(Error{
			Message: ErrCSVFileProcessingError,
		})
	}

	return PostUploadJSON200Response(Ok{
		Message: MsgUploadSuccessful,
	})
}

func ConvertReport(r payroll.PayrollReport) PayrollReport {
	empPayrolls := make([]WorkerPayrollBiWeek, 0, len(r.EmployeeReports))

	// sort by empId and payperiod
	sort.Slice(r.EmployeeReports, func(i, j int) bool {
		if r.EmployeeReports[i].EmployeeId < r.EmployeeReports[j].EmployeeId {
			return true
		} else if r.EmployeeReports[i].EmployeeId > r.EmployeeReports[j].EmployeeId {
			return false
		}

		return r.EmployeeReports[i].PayPeriod.StartDate.Before(r.EmployeeReports[j].PayPeriod.StartDate)
	})

	for _, empReport := range r.EmployeeReports {
		empPayrolls = append(empPayrolls, WorkerPayrollBiWeek{
			AmountPaid: fmt.Sprintf("$%.2f", empReport.AmountPaid),
			EmployeeID: uint64(empReport.EmployeeId),
			PayPeriod: struct {
				EndDate   *types.Date "json:\"end_date,omitempty\""
				StartDate *types.Date "json:\"start_date,omitempty\""
			}{
				StartDate: ConvertDate(empReport.PayPeriod.StartDate),
				EndDate:   ConvertDate(empReport.PayPeriod.EndDate),
			},
		})
	}

	return PayrollReport{
		EmployeeReports: empPayrolls,
	}
}

// ConvertDate func converts internal time object to openapi object
func ConvertDate(t time.Time) *openapi_types.Date {
	return &openapi_types.Date{
		Time: t,
	}
}

// ConvertWorkGroup func converts internal job group object to openapi object
func ConvertWorkGroup(s string) payroll.JobGroup {
	logJobGroup := payroll.GroupA
	if s == "B" {
		logJobGroup = payroll.GroupB
	}
	return logJobGroup
}

// ParseTime func converts time string to time object
func ParseTime(s string) (*time.Time, error) {
	parts := strings.Split(s, "/")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid date specified")
	}
	day, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid date specified: %v", err)
	}
	month, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid date specified: %v", err)
	}
	year, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid date specified: %v", err)
	}

	if day < 1 || day > 31 || month < 1 || month > 12 || len(parts[2]) != 4 {
		return nil, fmt.Errorf("invalid date specified")
	}

	// TODO: Check error
	t := time.Date(year, GetMonth(month), day, 0, 0, 0, 0, time.Local)
	return &t, nil
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
