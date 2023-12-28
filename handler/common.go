package handler

import (
	"github.com/joshinjohnson/wave-exercise/pkg/payroll"
)

type PayrollService interface {
	InsertLogs(filenameId int, logs []payroll.WorkLog) error
	GetReport(limit, offset uint64) (payroll.PayrollReport, error)
}

// API response messages
var (
	ErrHTTPForbidden           = "Forbidden"
	ErrHTTPInternalServerError = "Internal Server Error"
	ErrCSVFileProcessingError  = "Error reading csv file. Please upload a valid csv file"
	MsgUploadSuccessful        = "Upload successful"
	// ErrCSVFileAlreadyProcessedError = "Error reading csv file. Already processed file with same id"
)
