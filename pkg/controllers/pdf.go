package controllers

import (
	"Template/pkg/models"
	"Template/pkg/models/errors"
	"Template/pkg/models/response"
	"Template/pkg/utils/go-utils/database"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"

	wkhtml "github.com/SebastiaanKlippert/go-wkhtmltopdf"

	"github.com/gofiber/fiber/v2"
)

// initialize buffer to store html with data
var buf bytes.Buffer

// used for testing purposes
func HtmlTest(c *fiber.Ctx) error {
	// declare the html template
	// var temp *template.Template
	temp := template.Must(template.ParseFiles("pdf-templates/Cela-ePN-Template.html"))

	// initialize request_id, request key for template
	var request_id = "LOAN-5d710489-5404-4478-99f8-a328486520db"

	// declare variable with cela data struct
	data := &models.Cela_Data_In_Database{}

	//Get data from data source
	err := database.DBConn.Raw("SELECT * FROM cela_test_list_of_loan_applications WHERE request_id = ?", request_id).Find(&data).Error
	if err != nil {
		return c.JSON(response.ResponseModel{
			RetCode: "203",
			Message: "query error",
			Data:    err.Error(),
		})
	}

	// calculate total
	due, _ := strconv.ParseInt(data.Insterest_due, 10, 64)
	loan, _ := strconv.ParseInt(data.Loan_amount, 10, 64)

	total := due + loan

	data.Total = strconv.FormatInt(total, 10)

	// loan_amount to words
	if data.Loan_amount == "1000" {
		data.Amtword = "One Thousand"
	} else if data.Loan_amount == "2000" {
		data.Amtword = "Two Thousand"
	} else if data.Loan_amount == "3000" {
		data.Amtword = "Three Thousand"
	} else if data.Loan_amount == "4000" {
		data.Amtword = "Four Thousand"
	} else if data.Loan_amount == "5000" {
		data.Amtword = "Five Thousand"
	}

	// assure date_applied
	date_applied := data.Date_applied

	// format date_applied
	t, _ := time.Parse("2006-01-02", data.Date_applied)
	data.Date_applied = t.Format("January 02, 2006")

	// convert data
	term, _ := strconv.ParseInt(data.Term, 10, 64)

	// get data from LOS API
	// initialize request body
	reqBody := &models.LOS_Request_Body{
		Principal:     int(loan), // changeable
		Flatrate:      24,
		ProRateonHalf: 4,
		N:             int(term), // changaable
		Frequency:     50,
		DateReleased:  date_applied, // assured date format is "2023-03-28"
		MeetingDay:    6,            // ????
		DueDateType:   1,
		WithDST:       0,
		IsLumpSum:     0,
		IntComp:       1,
		GracePeriod:   0,
	}

	// Marshal JSON request body
	jsonReq, err := json.Marshal(reqBody)
	if err != nil {
		return c.JSON(response.ResponseModel{
			RetCode: "400",
			Message: "fail",
			Data:    err.Error(),
		})
	}

	// make request to LOS to calculate amortization
	req, err := http.NewRequest("POST", "https://loscbtest.cardmri.com:8444/LOSMobileAPI/ComputeAmortization", bytes.NewBuffer(jsonReq))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("charset", "utf-8")
	if err != nil {
		return c.JSON(response.ResponseModel{
			RetCode: "400",
			Message: "fail",
			Data:    err.Error(),
		})
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return c.JSON(response.ResponseModel{
			RetCode: "400",
			Message: "fail",
			Data:    err.Error(),
		})
	}

	defer resp.Body.Close()

	// read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return c.JSON(response.ResponseModel{
			RetCode: "400",
			Message: "fail",
			Data:    err.Error(),
		})
	}
	// fmt.Print(string(body))

	result := json.RawMessage(body)
	// fmt.Print(json.Valid(result))
	mapResult := make(map[string]interface{})
	if unmarErr := json.Unmarshal(result, &mapResult); unmarErr != nil {
		return c.JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Error",
			Data: errors.ErrorModel{
				Message:   "unmarshal error",
				IsSuccess: false,
				Error:     unmarErr,
			},
		})
	}

	// set start date and end date based from LOS
	dataslice := mapResult["amortization"].([]interface{})
	// start date
	amort_firstelem := dataslice[0].(map[string]interface{})
	// end data
	amort_lastelem := dataslice[len(dataslice)-1].(map[string]interface{})

	// set all data needed to be rendered in html
	allData := models.Cela_Data_To_Render{
		Data:         *data,
		Start_date:   amort_firstelem["dueDate2"].(string),
		End_date:     amort_lastelem["dueDate2"].(string),
		Amortization: mapResult["amortization"],
	}

	// reset buffer to ensure template and data is only executed once
	buf.Reset()

	// execute the template and data, store result in buffer
	err = temp.Execute(&buf, allData)
	if err != nil {
		return err
	}

	// for testing purposes
	// return c.JSON(response.ResponseModel{
	// 	RetCode: "200",
	// 	Message: "success",
	// 	Data:    amort,
	// })

	return c.SendString(buf.String())
}

// actual function to generate pdf
func PdfTest(c *fiber.Ctx) error {
	// pdf generator
	r, err := wkhtml.NewPDFGenerator()
	if err != nil {
		return err
	}

	// global options for pdf
	r.PageSize.Set(wkhtml.PageSizeLetter)

	// declare the html template
	// var temp *template.Template
	temp := template.Must(template.ParseFiles("pdf-templates/Cela-ePN-Template.html"))

	// initialize request id, request key for template
	var request_id = "LOAN-5d710489-5404-4478-99f8-a328486520db"

	// declare variable with cela data struct
	data := &models.Cela_Data_In_Database{}

	//Get data from data source, test db
	err = database.DBConn.Raw("SELECT * FROM cela_test_list_of_loan_applications WHERE request_id = ?", request_id).Find(&data).Error
	if err != nil {
		return c.JSON(response.ResponseModel{
			RetCode: "203",
			Message: "query error",
			Data:    err.Error(),
		})
	}

	// formatting data
	// calculate total
	due, _ := strconv.ParseInt(data.Insterest_due, 10, 64)
	loan, _ := strconv.ParseInt(data.Loan_amount, 10, 64)

	total := due + loan

	data.Total = strconv.FormatInt(total, 10)

	// loan_amount to words
	if data.Loan_amount == "1000" {
		data.Amtword = "One Thousand"
	} else if data.Loan_amount == "2000" {
		data.Amtword = "Two Thousand"
	} else if data.Loan_amount == "3000" {
		data.Amtword = "Three Thousand"
	} else if data.Loan_amount == "4000" {
		data.Amtword = "TFour Thousand"
	} else if data.Loan_amount == "5000" {
		data.Amtword = "Five Thousand"
	}

	// assure date_applied
	date_applied := data.Date_applied

	// format date_applied
	t, _ := time.Parse("2006-01-02", data.Date_applied)
	data.Date_applied = t.Format("January 02, 2006")

	// convert data
	term, _ := strconv.ParseInt(data.Term, 10, 64)

	// get data from LOS API
	// initialize request body
	reqBody := &models.LOS_Request_Body{
		Principal:     int(loan), // changeable
		Flatrate:      24,
		ProRateonHalf: 4,
		N:             int(term), // changaable
		Frequency:     50,
		DateReleased:  date_applied, // ????
		MeetingDay:    6,            // ????
		DueDateType:   1,
		WithDST:       0,
		IsLumpSum:     0,
		IntComp:       1,
		GracePeriod:   0,
	}

	// Marshal JSON request body
	jsonReq, err := json.Marshal(reqBody)
	if err != nil {
		return c.JSON(response.ResponseModel{
			RetCode: "400",
			Message: "fail",
			Data:    err.Error(),
		})
	}

	// make request to LOS to calculate amortization
	req, err := http.NewRequest("POST", "https://loscbtest.cardmri.com:8444/LOSMobileAPI/ComputeAmortization", bytes.NewBuffer(jsonReq))
	if err != nil {
		return c.JSON(response.ResponseModel{
			RetCode: "400",
			Message: "fail",
			Data:    err.Error(),
		})
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return c.JSON(response.ResponseModel{
			RetCode: "400",
			Message: "fail",
			Data:    err.Error(),
		})
	}

	defer resp.Body.Close()

	// read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return c.JSON(response.ResponseModel{
			RetCode: "400",
			Message: "fail",
			Data:    err.Error(),
		})
	}

	result := json.RawMessage(body)
	mapResult := make(map[string]interface{})
	if unmarErr := json.Unmarshal(result, &mapResult); unmarErr != nil {
		return c.JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Error",
			Data: errors.ErrorModel{
				Message:   "unmarshal error",
				IsSuccess: false,
				Error:     unmarErr,
			},
		})
	}

	// set start date and end date based from LOS
	dataslice := mapResult["amortization"].([]interface{})
	// start date
	amort_firstelem := dataslice[0].(map[string]interface{})
	// end data
	amort_lastelem := dataslice[len(dataslice)-1].(map[string]interface{})

	// set all data needed to be rendered in html
	allData := models.Cela_Data_To_Render{
		Data:         *data,
		Start_date:   amort_firstelem["dueDate2"].(string),
		End_date:     amort_lastelem["dueDate2"].(string),
		Amortization: mapResult["amortization"],
	}

	// reset buffer to ensure template and data is only executed once
	buf.Reset()

	// execute the template and data, store result in buffer
	err = temp.Execute(&buf, allData)
	if err != nil {
		return err
	}

	// use buffer instead
	// convert html to string
	str := buf.String()

	// r.AddPage(wkhtml.NewPageReader(strings.NewReader(str)))

	// set page 1
	// code automatically generates new pages
	page1 := wkhtml.NewPageReader(strings.NewReader(str))

	// set page 1 options
	page1.Allow.Set("/Users/fdsap-gatan/Projects/Pdf-Generator/pdf-templates")
	page1.EnableLocalFileAccess.Set(true)

	// add page
	r.AddPage(page1)

	// Create PDF document in internal buffer
	err = r.Create()
	if err != nil {
		return err
	}

	//Your Pdf Name
	err = r.WriteFile("./test.pdf")
	if err != nil {
		return err
	}

	return c.SendString("success")
}
