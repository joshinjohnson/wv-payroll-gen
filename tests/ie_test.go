package tests

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/joshinjohnson/wave-exercise/handler"
	"github.com/joshinjohnson/wave-exercise/pkg/db"
	"github.com/joshinjohnson/wave-exercise/pkg/payroll"
	"github.com/sirupsen/logrus"
)

var (
	payAPIhandler handler.PayrollHandler
	dbW           *db.DbWrapper
)

func TestUploadCSV(t *testing.T) {
	body, boundary := createCSVRequest()
	req, err := http.NewRequest("POST", "/upload", body)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "multipart/form-data; boundary="+boundary)
	rr := httptest.NewRecorder()

	payAPIhandler, dbW = setupHandler()
	defer dbW.DB.Close()
	payAPIhandler.PostUpload(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("TestUploadCSV returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestGetPayrollReport(t *testing.T) {
	req, err := http.NewRequest("GET", "/report", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	payAPIhandler, dbW = setupHandler()
	defer dbW.DB.Close()
	payAPIhandler.GetReport(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("TestGetPayrollReport returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func createCSVRequest() (*bytes.Buffer, string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	file, err := os.Open("./fixtures/time-report-42.csv")
	if err != nil {
		logrus.Errorf("Error opening file: %v", err)
		return nil, ""
	}
	defer file.Close()

	part, err := writer.CreateFormFile("file", "time-report-42.csv")
	if err != nil {
		logrus.Errorf("Error creating form file: %v", err)
		return nil, ""
	}

	_, err = io.Copy(part, file)
	if err != nil {
		logrus.Errorf("Error copying file: %v", err)
		return nil, ""
	}

	boundary := writer.Boundary()
	writer.Close()

	return body, boundary
}

func setupHandler() (handler.PayrollHandler, *db.DbWrapper) {
	dbConfig := map[string]string{
		"user":     "user",
		"password": "pass@123",
		"hostname": "localhost",
		"port":     "5432",
		"db-name":  "payroll",
		"schema":   "public",
		"ssl-mode": "disable",
	}

	dbW, err := db.NewDbWrapper(context.TODO(), dbConfig)
	if err != nil {
		logrus.Errorf("error while setting up db client: %v", err)
		os.Exit(1)
	}

	payrollService := payroll.NewPayrollService(dbW)
	payrollHandler := handler.NewPayrollHandler(payrollService)
	return payrollHandler, dbW
}
