package payroll

import "fmt"

var (
	ErrWorkLogCreate  = fmt.Errorf("error while inserting job log/s")
	ErrFileIdExists   = fmt.Errorf("file id already processed")
	ErrReportGenerate = fmt.Errorf("error while generating the report")
)
