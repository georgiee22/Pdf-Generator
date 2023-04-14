package models

// data struct for CARMELA
// declare data struct to be rendered in template
type Carmela_Data struct {
	Pnum        string `json:"pnum"`
	Date        string `json:"date"`
	Name        string `json:"name"`
	Spouse_name string `json:"spousename"`
	Addresses   string `json:"addresses"`
	Prin        string `json:"prin"`
	M_term      string `json:"mterm"`
	Nomflatrate string `json:"nomflatrate"`
	Eir         string `json:"eir"`
	Termdays    string `json:"termdays"`
	Elrf        string `json:"elrf"`
	Term        string `json:"term"`
	Weekly_due  string `json:"weeklydue"`
	Amtword     string `json:"amtword"`
	Fpay        string `json:"fpay"`
	Enddate     string `json:"enddate"`
	Settle      string `json:"settle"`
}

// data struct for CELA-ePN-Template
// declare data struct to be rendered in template
type Cela_Data_To_Render struct {
	Data         Cela_Data_In_Database `json:"celadata"`
	Start_date   string                `json:"startdate"`    // aka Fpay
	End_date     string                `json:"enddate"`      // aka Enddate
	Amortization any                   `json:"amortization"` // list of amortization
}

type Cela_Data_In_Database struct {
	Pn_no              string `json:"pnno"`       // aka Pnum
	Name               string `json:"name"`       // name
	Addresses          string `json:"addresses"`  // ????
	Loan_amount        string `json:"loanamount"` // aka principal
	Contractual_rate   string `json:"contractualrate"`
	Eir                string `json:"eir"`
	Insterest_due      string `json:"Interest_due"` // aka total interest
	Elrf               string `json:"elrf"`
	Term               string `json:"term"`              // term
	Total              string `json:"total"`             // calculated with interest_due+Loan_amount
	Amtword            string `json:"amtword"`           // loan_amount converted to words
	Date_applied       string `json:"dateapplied"`       // aka sysdate
	Customer_number    string `json:"customernumber"`    // aka customer_id
	Settlement_account string `json:"settlementaccount"` // aka account number ????
}

// declare request body for LOS
type LOS_Request_Body struct {
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
