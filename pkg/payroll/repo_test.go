package payroll_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	internaldb "github.com/joshinjohnson/wave-exercise/pkg/db"
	"github.com/joshinjohnson/wave-exercise/pkg/payroll"
	"github.com/stretchr/testify/assert"
)

var (
	selectCols              = "employee_id, log_date, log_hours, job_group"
	insertCols              = "employee_id, log_date, log_hours, job_group, updated_ts"
	insertColsCount         = 5
	selectJobGroupRateQuery = "select job_group, rate from jobgroup_rate;"
	selectLogsQuery         = "select " + selectCols + " from worklog order by log_date limit $1 offset $2;"
	insertFileIdQuery       = "insert into processed_files values ($1);"
	insertLogsQuery         = "insert into worklog (" + insertCols + ") values <replace> returning id;"
	timeVal                 = time.Now()
)

func TestGetJobGroupRates(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock database: %v", err)
	}
	defer db.Close()

	repo := payroll.NewPayrollRepository(&internaldb.DbWrapper{
		DB: db,
	})

	expectedRates := []payroll.JobGroupRate{
		{JobGroup: "A", Rate: 30},
		{JobGroup: "B", Rate: 20},
	}

	rows := sqlmock.NewRows([]string{"job_group", "rate"}).
		AddRow("A", 30).
		AddRow("B", 20)

	mock.ExpectQuery("select job_group, rate from jobgroup_rate;").WillReturnRows(rows)

	actualRates, err := repo.GetJobGroupRates()

	assert.NoError(t, err)
	assert.Equal(t, expectedRates, actualRates)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetJobGroupRates_ErrorOnQuery(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock database: %v", err)
	}
	defer db.Close()

	repo := payroll.NewPayrollRepository(&internaldb.DbWrapper{
		DB: db,
	})

	expectedError := fmt.Errorf("query error")
	mock.ExpectQuery("select job_group, rate from jobgroup_rate;").WillReturnError(expectedError)

	_, err = repo.GetJobGroupRates()

	assert.Error(t, err)
	assert.EqualError(t, err, expectedError.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetJobGroupRates_ErrorOnScan(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock database: %v", err)
	}
	defer db.Close()

	repo := payroll.NewPayrollRepository(&internaldb.DbWrapper{
		DB: db,
	})

	rows := sqlmock.NewRows([]string{"job_group", "rate"}).
		AddRow("A", 30).
		AddRow("B", "invalid")

	mock.ExpectQuery("select job_group, rate from jobgroup_rate;").WillReturnRows(rows)

	_, err = repo.GetJobGroupRates()

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateN(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock database: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	tx, _ := db.Begin()
	repo := payroll.NewPayrollRepository(&internaldb.DbWrapper{
		DB: db,
		Tx: tx,
	})

	expectedLogs := []payroll.WorkLog{
		{EmployeeId: 1, Date: timeVal, HoursLogged: 8, JobGroup: "A"},
		{EmployeeId: 2, Date: timeVal, HoursLogged: 6, JobGroup: "B"},
	}

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2)

	mock.ExpectQuery("insert into worklog *").
		WithArgs(1, timeVal, 8.0, "A", sqlmock.AnyArg(), 2, timeVal, 6.0, "B", sqlmock.AnyArg()).
		WillReturnRows(rows)
	mock.ExpectCommit()

	_, err = repo.CreateN(expectedLogs)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateN_ErrorOnQuery(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock database: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	tx, _ := db.Begin()
	repo := payroll.NewPayrollRepository(&internaldb.DbWrapper{
		DB: db,
		Tx: tx,
	})

	expectedLogs := []payroll.WorkLog{
		{EmployeeId: 1, Date: timeVal, HoursLogged: 8, JobGroup: "A"},
		{EmployeeId: 2, Date: timeVal, HoursLogged: 6, JobGroup: "B"},
	}

	expectedError := fmt.Errorf("query error")
	mock.ExpectQuery("insert into worklog (.+) values (.+) returning id;").
		WithArgs(1, timeVal, 8.0, "A", sqlmock.AnyArg(), 2, timeVal, 6.0, "B", sqlmock.AnyArg()).
		WillReturnError(expectedError)

	_, err = repo.CreateN(expectedLogs)

	assert.Error(t, err)
	assert.EqualError(t, err, expectedError.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateN_ErrorOnScan(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock database: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	tx, _ := db.Begin()
	repo := payroll.NewPayrollRepository(&internaldb.DbWrapper{
		DB: db,
		Tx: tx,
	})

	expectedLogs := []payroll.WorkLog{
		{EmployeeId: 1, Date: timeVal, HoursLogged: 8, JobGroup: "A"},
		{EmployeeId: 2, Date: timeVal, HoursLogged: 6, JobGroup: "B"},
	}

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(1).
		AddRow("invalid")

	mock.ExpectQuery("insert into worklog (.+) values (.+) returning id;").
		WithArgs(1, timeVal, 8.0, "A", sqlmock.AnyArg(), 2, timeVal, 6.0, "B", sqlmock.AnyArg()).
		WillReturnRows(rows)

	_, err = repo.CreateN(expectedLogs)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPlaceholderGen(t *testing.T) {
	query := "insert into worklog (employee_id, log_date, log_hours) values <replace>;"
	argsLen := 3
	startIdx := 1

	result := payroll.PlaceholderGen(query, argsLen, startIdx)
	expectedResult := "insert into worklog (employee_id, log_date, log_hours) values $1,$2,$3;"
	assert.Equal(t, expectedResult, result)
}

func TestPlaceholderGen_NoReplace(t *testing.T) {
	query := "select * from table where condition = true;"
	argsLen := 3
	startIdx := 1

	result := payroll.PlaceholderGen(query, argsLen, startIdx)
	assert.Equal(t, query, result)
}

func TestPlaceholderGenBulk(t *testing.T) {
	query := "insert into worklog (employee_id, log_date, log_hours) values <replace>;"
	argsLen := 3
	totalArgs := 2
	startIdx := 1

	result, err := payroll.PlaceholderGenBulk(query, argsLen, totalArgs, startIdx)
	expectedResult := "insert into worklog (employee_id, log_date, log_hours) values ($1,$2,$3),($4,$5,$6);"
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
}

func TestPlaceholderGenBulk_InvalidArgsLen(t *testing.T) {
	query := "insert into worklog (employee_id, log_date, log_hours) values <replace>;"
	argsLen := 0
	totalArgs := 2
	startIdx := 1

	result, err := payroll.PlaceholderGenBulk(query, argsLen, totalArgs, startIdx)
	assert.Error(t, err)
	assert.Equal(t, query, result)
}

func TestPlaceholderGenBulk_InvalidTotalArgs(t *testing.T) {
	query := "insert into worklog (employee_id, log_date, log_hours) values <replace>;"
	argsLen := 3
	totalArgs := 0
	startIdx := 1

	result, err := payroll.PlaceholderGenBulk(query, argsLen, totalArgs, startIdx)
	assert.Error(t, err)
	assert.Equal(t, query, result)
}

func TestPlaceholderGenBulk_NoReplace(t *testing.T) {
	query := "select * from table where condition = true;"
	argsLen := 3
	totalArgs := 2
	startIdx := 1

	result, err := payroll.PlaceholderGenBulk(query, argsLen, totalArgs, startIdx)
	assert.Error(t, err)
	assert.Equal(t, query, result)
}

func TestFlattenLogInsertArgs(t *testing.T) {
	params := []payroll.WorkLog{
		{EmployeeId: 1, Date: time.Now(), HoursLogged: 8, JobGroup: "A"},
		{EmployeeId: 2, Date: time.Now(), HoursLogged: 6, JobGroup: "B"},
	}

	result := payroll.FlattenLogInsertArgs(params)
	assert.Equal(t, 10, len(result))
}

func TestFlattenLogInsertArgs_EmptyParams(t *testing.T) {
	var params []payroll.WorkLog

	result := payroll.FlattenLogInsertArgs(params)
	assert.Empty(t, result)
}
