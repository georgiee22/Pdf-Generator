package controllers

import (
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
	var temp *template.Template
	temp = template.Must(template.ParseFiles("pdf-templates/Cela-ePN-Template.html"))

	// // data struct for CARMELA
	// // declare data struct to be rendered in template
	// type carmela_data struct {
	// 	Pnum        string `json:"pnum"`
	// 	Date        string `json:"date"`
	// 	Name        string `json:"name"`
	// 	Spouse_name string `json:"spousename"`
	// 	Addresses   string `json:"addresses"`
	// 	Prin        string `json:"prin"`
	// 	M_term      string `json:"mterm"`
	// 	Nomflatrate string `json:"nomflatrate"`
	// 	Eir         string `json:"eir"`
	// 	Termdays    string `json:"termdays"`
	// 	Elrf        string `json:"elrf"`
	// 	Term        string `json:"term"`
	// 	Weekly_due  string `json:"weeklydue"`
	// 	Amtword     string `json:"amtword"`
	// 	Fpay        string `json:"fpay"`
	// 	Enddate     string `json:"enddate"`
	// 	Settle      string `json:"settle"`
	// }

	// data struct for CELA
	// declare data struct to be rendered in template
	type cela_data struct {
		Pn_no              string `json:"pnno"` // aka Pnum
		Name               string `json:"name"` // name
		Addresses          string `json:"addresses"`
		Loan_amount        string `json:"loanamount"` // aka principal
		Contractual_rate   string `json:"contractrualrate"`
		Eir                string `json:"eir"`
		Insterest_due      string `json:"Interest_due"` // aka total interest
		Elrf               string `json:"elrf"`
		Term               string `json:"term"`              // term
		Total              string `json:"total"`             // calculated with interest_due+Loan_amount
		Amtword            string `json:"amtword"`           // loan_amount converted to words
		Start_date         string `json:"startdate"`         // aka Fpay
		End_date           string `json:"enddate"`           // aka Enddate
		Date_applied       string `json:"dateapplied"`       // aka sysdate
		Customer_number    string `json:"customernumber"`    // aka customer_id
		Settlement_account string `json:"settlementaccount"` // aka account number ????
	}

	// initialize request_id, request key for template
	var request_id = "LOAN-5d710489-5404-4478-99f8-a328486520db"

	// declare variable with cela data struct
	var data cela_data

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
		data.Amtword = "TFour Thousand"
	} else if data.Loan_amount == "5000" {
		data.Amtword = "Five Thousand"
	}

	// format date_applied
	t, _ := time.Parse("2006-01-02", data.Date_applied)
	data.Date_applied = t.Format("January 02, 2006")

	// declare request body for LOS
	type LOSRequestBody struct {
		Principal     int    `json:"principal"`
		Flatrate      int    `json:"flatRate"`
		ProRateonHalf int    `json:"proRateonHalf"`
		N             int    `json:"n"`
		Frequency     int    `json:"frequency"`
		DateReleased  string `json:"dateReleased"`
		MeetingDay    int    `json:"meetingDay"`
		DueDateType   int    `json:"dueDateType"`
		WithDST       int    `json:"withDST"`
		IsLumpSum     int    `json:"isLumpSum"`
		IntComp       int    `json:"intComp"`
		GracePeriod   int    `json:"gracePeriod"`
	}

	// convert data
	term, _ := strconv.ParseInt(data.Term, 10, 64)

	// get data from LOS API
	// initialize request body
	reqBody := &LOSRequestBody{
		Principal:     int(loan), // changeable
		Flatrate:      24,
		ProRateonHalf: 4,
		N:             int(term), // changaable
		Frequency:     50,
		DateReleased:  data.Date_applied, // ????
		MeetingDay:    6,                 // ????
		DueDateType:   1,
		WithDST:       0,
		IsLumpSum:     0,
		IntComp:       1,
		GracePeriod:   0,
	}

	// declare amortization struct
	type Amor struct {
		Num       int     `json:"num"`
		Duedate   string  `json:"dueDate"`
		DuePrin   float32 `json:"duePrin"`
		DueInt    float32 `json:"dueInt"`
		BalPrin   float32 `json:"balPrin"`
		BalInt    float32 `json:"balInt"`
		ExcessInt float32 `json:"excessInt"`
		DueDate2  string  `json:"dueDate2"`
		DueTotal  float32 `json:"dueTotal"`
	}

	type AmorResponse []Amor

	// declare response body
	type LOSResponseBody struct {
		Principal           json.Number  `json:"principal"`
		Flatrate            float32      `json:"flatRate"`
		ProRateonHalf       float32      `json:"proRateonHalf"`
		N                   int          `json:"n"`
		DateReleased        string       `json:"dateReleased"`
		MeetingDay          int          `json:"meetingDay"`
		DueDateType         int          `json:"dueDateType"`
		WithDST             int          `json:"withDST"`
		IsLumpSum           int          `json:"isLumpSum"`
		GracePeriod         int          `json:"gracePeriod"`
		Frequency           string       `json:"frequency"`
		InterestComputation string       `json:"interestComputation"`
		Rate                float64      `json:"rate"`
		DueAmt              float64      `json:"dueAmt"`
		AddonPeriondInt     float64      `json:"addonPeriodInt"`
		DateReleased2       string       `json:"dateReleased2"`
		BspEffective        float64      `json:"bspEffective"`
		Okay                bool         `json:"okay"`
		Amortization        AmorResponse `json:"amortization"`
		FirstPayment        string       `json:"firstPayment"`
		TotalInterest       float64      `json:"totalInterest"`
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
	//fmt.Print(body)

	result := json.RawMessage(body)
	var mapResult LOSResponseBody
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

	// reset buffer to ensure template and data is only executed once
	buf.Reset()

	// execute the template and data, store result in buffer
	err = temp.Execute(&buf, data)
	if err != nil {
		return err
	}

	return c.JSON(response.ResponseModel{
		RetCode: "200",
		Message: "success",
		Data:    mapResult,
	})

	// return c.SendString(buf.String())
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
	var temp *template.Template
	temp = template.Must(template.ParseFiles("pdf-templates/Cela-ePN-Template.html"))

	// // data struct for CARMELA
	// // declare data struct to be rendered in template
	// type carmela_data struct {
	// 	Pnum        string `json:"pnum"`
	// 	Date        string `json:"date"`
	// 	Name        string `json:"name"`
	// 	Spouse_name string `json:"spousename"`
	// 	Addresses   string `json:"addresses"`
	// 	Prin        string `json:"prin"`
	// 	M_term      string `json:"mterm"`
	// 	Nomflatrate string `json:"nomflatrate"`
	// 	Eir         string `json:"eir"`
	// 	Termdays    string `json:"termdays"`
	// 	Elrf        string `json:"elrf"`
	// 	Term        string `json:"term"`
	// 	Weekly_due  string `json:"weeklydue"`
	// 	Amtword     string `json:"amtword"`
	// 	Fpay        string `json:"fpay"`
	// 	Enddate     string `json:"enddate"`
	// 	Settle      string `json:"settle"`
	// }

	// data struct for CELA
	// declare data struct to be rendered in template
	// data source is postgre cela_test
	type cela_data struct {
		Pn_no              string `json:"pnno"` // aka Pnum
		Name               string `json:"name"` // name
		Addresses          string `json:"addresses"`
		Loan_amount        string `json:"loanamount"` // aka principal
		Contractual_rate   string `json:"contractrualrate"`
		Eir                string `json:"eir"`
		Insterest_due      string `json:"Interest_due"` // aka total interest ????
		Elrf               string `json:"elrf"`
		Term               string `json:"term"`              // term
		Total              string `json:"total"`             // calculated with interest_due+Loan_amount
		Amtword            string `json:"amtword"`           // loan_amount converted to words
		Start_date         string `json:"startdate"`         // aka Fpay
		End_date           string `json:"enddate"`           // aka Enddate
		Date_applied       string `json:"dateapplied"`       // aka sysdate
		Customer_number    string `json:"customernumber"`    // aka customer_id
		Settlement_account string `json:"settlementaccount"` // aka account number ????
	}

	// initialize cid, request key for template
	var request_id = "LOAN-5d710489-5404-4478-99f8-a328486520db"

	// declare variable with cela data struct
	var data cela_data

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

	// format date_applied
	t, _ := time.Parse("2006-01-02", data.Date_applied)
	data.Date_applied = t.Format("January 02, 2006")

	type LOSRequestBody struct {
		Principal     int    `json:"principal"`
		Flatrate      int    `json:"flatRate"`
		ProRateonHalf int    `json:"proRateonHalf"`
		N             int    `json:"n"`
		Frequency     int    `json:"frequency"`
		DateReleased  string `json:"dateReleased"`
		MeetingDay    int    `json:"meetingDay"`
		DueDateType   int    `json:"dueDateType"`
		WithDST       int    `json:"withDST"`
		IsLumpSum     int    `json:"isLumpSum"`
		IntComp       int    `json:"intComp"`
		GracePeriod   int    `json:"gracePeriod"`
	}

	// convert data
	term, _ := strconv.ParseInt(data.Term, 10, 64)

	// get data from LOS API
	// initialize request body
	reqBody := &LOSRequestBody{
		Principal:     int(loan), // changeable
		Flatrate:      24,
		ProRateonHalf: 4,
		N:             int(term), // changaable
		Frequency:     50,
		DateReleased:  data.Date_applied, // ????
		MeetingDay:    6,                 // ????
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

	// reset buffer to ensure template and data is only executed once
	buf.Reset()

	// execute the template and data, store result in buffer
	err = temp.Execute(&buf, data)
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
	page1.Allow.Set("/Users/g.tan/Projects/Pdf-Generator/Go_Template/pdf-templates")
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
