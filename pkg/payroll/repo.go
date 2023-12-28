package payroll

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/joshinjohnson/wave-exercise/pkg/db"
	"github.com/sirupsen/logrus"
)

const (
	worklogTable   = "worklog"
	jobgroupTable  = "jobgroup_rate"
	processedTable = "processed_files"
)

var (
	selectCols              = "employee_id, log_date, log_hours, job_group"
	insertCols              = "employee_id, log_date, log_hours, job_group, updated_ts"
	insertColsCount         = 5
	selectJobGroupRateQuery = "select job_group, rate from " + jobgroupTable + ";"
	selectLogsQuery         = "select " + selectCols + " from " + worklogTable + " order by log_date limit $1 offset $2;"
	insertFileIdQuery       = "insert into " + processedTable + " values ($1);"
	insertLogsQuery         = "insert into " + worklogTable + " (" + insertCols + ") values <replace> returning id;"
)

type payrollRepository struct {
	dbW *db.DbWrapper
}

func NewPayrollRepository(dbW *db.DbWrapper) *payrollRepository {
	return &payrollRepository{
		dbW: dbW,
	}
}

func (r payrollRepository) GetJobGroupRates() ([]JobGroupRate, error) {
	gr := make([]JobGroupRate, 0)

	rows, err := r.dbW.DB.Query(selectJobGroupRateQuery)
	if err != nil {
		logrus.Errorf(fmt.Sprintf("error while fetching group rates: %v", err))
		return gr, err
	}

	defer rows.Close()

	for rows.Next() {
		var j JobGroupRate

		if err := rows.Scan(&j.JobGroup, &j.Rate); err != nil {
			logrus.Error(fmt.Sprintf("unable to scan db rows: %v", err))
			return gr, err
		}

		gr = append(gr, j)
	}

	return gr, nil
}

func (r payrollRepository) Get(limit, offset uint64) ([]WorkLog, error) {
	wl := make([]WorkLog, 0)

	rows, err := r.dbW.DB.Query(selectLogsQuery, limit, offset)
	if err != nil {
		logrus.Errorf(fmt.Sprintf("error while fetching work logs: %v", err))
		return wl, err
	}

	defer rows.Close()

	for rows.Next() {
		var j WorkLog

		if err := rows.Scan(&j.EmployeeId, &j.Date, &j.HoursLogged, &j.JobGroup); err != nil {
			logrus.Error(fmt.Sprintf("unable to scan db rows: %v", err))
			return wl, err
		}

		wl = append(wl, j)
	}

	return wl, nil
}

func (r payrollRepository) InsertFileId(id int) error {
	if r.dbW.Tx == nil {
		logrus.Errorf("not running in tx, stopping")
		return fmt.Errorf("no tx running")
	}

	var rows *sql.Rows
	var err error
	if rows, err = r.dbW.Tx.Query(insertFileIdQuery, id); err != nil {
		logrus.Infof("error while inserting file id: %v", err)
		r.dbW.Tx.Rollback()
		return fmt.Errorf("error while inserting file id: %v", err)
	}
	defer rows.Close()

	return nil
}

func (r payrollRepository) CreateN(js []WorkLog) ([]uint64, error) {
	var ids []uint64

	query, err := PlaceholderGenBulk(insertLogsQuery, insertColsCount, len(js), 1)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	if r.dbW.Tx == nil {
		logrus.Errorf("not running in tx, stopping")
		return []uint64{}, fmt.Errorf("no tx running")
	}

	rows, err := r.dbW.Tx.Query(query, FlattenLogInsertArgs(js)...)
	if err != nil {
		logrus.Errorf(fmt.Sprintf("unable to insert logs: %v", err))
		r.dbW.Tx.Rollback()
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id uint64

		err := rows.Scan(&id)
		if err != nil {
			logrus.Errorf(fmt.Sprintf("unable to scan db rows: %v", err))
			r.dbW.Tx.Rollback()
			return nil, err
		}

		ids = append(ids, id)
	}

	logrus.Debugf(fmt.Sprintf("created logs with id: %d", ids))

	if err := r.dbW.Tx.Commit(); err != nil {
		r.dbW.Tx.Rollback()
		return nil, err
	}

	return ids, nil
}

// PlaceholderGen generates argument part of insert query
// Example: `($1, $2)`
func PlaceholderGen(query string, argsLen int, startIdx int) string {
	if argsLen == 0 {
		return ""
	}
	if !strings.Contains(query, "<replace>") {
		return query
	}

	var res strings.Builder
	n := argsLen + startIdx

	for i := startIdx; i < n; i++ {
		res.WriteByte('$')
		res.WriteString(strconv.Itoa(i))
		if i < n-1 {
			res.WriteByte(',')
		}
	}

	return strings.Replace(query, "<replace>", res.String(), 1)
}

// PlaceholderGenBulk generates argument part of bulk insert query
// Example: `($1, $2), ($3, $4)`
func PlaceholderGenBulk(query string, argsLen int, totalArgs int, startIdx int) (string, error) {
	if argsLen == 0 || totalArgs == 0 {
		return query, fmt.Errorf("argsLen or totalArgs is invalid")
	}
	if !strings.Contains(query, "<replace>") {
		return query, fmt.Errorf("query doesn't contain <replace> placeholder")
	}

	var res strings.Builder

	// generate query `(<replace>), (<replace>)`
	for i := 0; i < totalArgs; i++ {
		res.WriteByte('(')
		res.WriteString(PlaceholderGen("<replace>", argsLen, startIdx))
		res.WriteByte(')')
		if i < totalArgs-1 {
			res.WriteByte(',')
		}
		startIdx += argsLen
	}

	return strings.Replace(query, "<replace>", res.String(), 1), nil
}

// "employee_id, log_date, log_hours, job_group"
func FlattenLogInsertArgs(params []WorkLog) []any {
	r := make([]any, 0)
	now := time.Now()

	for _, param := range params {
		r = append(r, param.EmployeeId)
		r = append(r, param.Date)
		r = append(r, param.HoursLogged)
		r = append(r, param.JobGroup)
		r = append(r, now)
	}

	return r
}
