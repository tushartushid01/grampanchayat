package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"grampanchayat/database"
	"grampanchayat/database/helper"
	"grampanchayat/models"
	"grampanchayat/utilities"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var JwtKey = []byte("secret_key")

var accountSid = os.Getenv("TWILIO_ACCOUNT_SID")
var authToken = os.Getenv("TWILIO_AUTH_TOKEN")
var VerifyServiceSid = os.Getenv("VerifyServiceSid")
var SMSAuthorizationKey = "WOFC7oiwS1TdnzbDMcsIx4hUemLuGHZjr8RtBpyV9v03qJklQPorFvS7I42KypXh03dUjOq6RBi9DwH1"

const defaultLimit = 100

var client = twilio.NewRestClientWithParams(twilio.ClientParams{
	Username: accountSid,
	Password: authToken,
})

func SendOTP(w http.ResponseWriter, r *http.Request) {
	var sendOTP models.SendOTP

	decoderErr := utilities.Decoder(r, &sendOTP)
	if decoderErr != nil {
		utilities.HandlerError(w, http.StatusBadRequest, "SendOTP: Decoder error:", decoderErr)
		return
	}

	if sendOTP.Phone == "" {
		utilities.HandlerError(w, http.StatusBadRequest, "phone number cannot be empty", errors.New("SendOTP: phone no cannot be empty"))
		return
	}

	err := SendSms(sendOTP.Phone, r)
	if err != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "SendOTP: Unable to send otp. %v", err)
		return
	}
}

func SendSms(toPhone string, req *http.Request) error {
	var otp string
	if os.Getenv("BRANCH") == "DEV" {
		otp = "9999"
	} else {
		var randomCodes = [...]byte{
			'1', '2', '3', '4', '5', '6', '7', '8', '9', '0',
		}

		r := rand.New(rand.NewSource(time.Now().UnixNano()))

		pwd := make([]byte, 4)

		for j := 0; j < 4; j++ {
			index := r.Int() % len(randomCodes)
			pwd[j] = randomCodes[index]
		}
		otp = string(pwd)

		message := "This%20is%20your%20OTP : "

		url := fmt.Sprintf("https://www.fast2sms.com/dev/bulkV2?sender_id=TXTIND&message=%v%v&route=v3&numbers=%v", message, otp, toPhone)

		payload := strings.NewReader(otp)

		client := &http.Client{}
		req, err := http.NewRequest(http.MethodPost, url, payload)
		if err != nil {
			logrus.Printf("SendSms: unable to create request. %v", err)
			return err
		}
		req.Header.Add("Authorization", SMSAuthorizationKey)
		res, err := client.Do(req)
		if err != nil {
			logrus.Printf("SendSms: unable to get response. %v", err)
			return err
		}
		defer res.Body.Close()

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			logrus.Printf("SendSms: unable to read response bidy. %v", err)
			return err
		}
		fmt.Println(string(body))
	}
	err := helper.AddOtp(toPhone, otp)
	if err != nil {
		logrus.Printf("SendSms: cannot send sms:%v", err)
		return err
	}

	return nil
}

func LoginWithOTP(w http.ResponseWriter, r *http.Request) {
	var loginNumber models.LoginWithOTP

	decoderErr := utilities.Decoder(r, &loginNumber)
	if decoderErr != nil {
		utilities.HandlerError(w, http.StatusBadRequest, "LoginWithOTP: Decoder error:", decoderErr)
		return
	}

	if loginNumber.OTP == "" {
		utilities.HandlerError(w, http.StatusBadRequest, "OTP cannot be empty:", errors.New("OTP cannot be empty"))
		return
	}

	if loginNumber.Phone == "" {
		utilities.HandlerError(w, http.StatusBadRequest, "mobile number cannot be empty:", errors.New("mobile number cannot be empty"))
		return
	}

	storedOTP, err := helper.FetchOTP(loginNumber.Phone)
	if err != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "FetchOTP: cannot get otp:", err)
		return
	}

	if storedOTP != loginNumber.OTP {
		utilities.HandlerError(w, http.StatusBadRequest, "invalid otp", errors.New("invalid otp"))
		return
	}

	userCredentials, fetchErr := helper.FetchUserIDAndRole(loginNumber.Phone)
	if fetchErr != nil {
		if fetchErr == sql.ErrNoRows {
			_, err := w.Write([]byte("ERROR: Wrong details"))
			if err != nil {
				return
			}
			utilities.HandlerError(w, http.StatusBadRequest, "FetchUserIDAndRole:", fetchErr)
			return
		}
		utilities.HandlerError(w, http.StatusInternalServerError, "FetchUserIDAndRole:", fetchErr)
		return
	}

	expiresAt := time.Now().Add(60 * time.Hour)

	claims := &models.Claims{
		ID:   userCredentials.ID,
		Role: userCredentials.Role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresAt.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(JwtKey)
	if err != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "TokenString: cannot create token string:", err)
		return
	}

	err = helper.CreateSession(claims)
	if err != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "LoginWithOTP: CreateSession:", err)
		return
	}

	userOutboundData := make(map[string]interface{})

	userOutboundData["token"] = tokenString

	err = utilities.Encoder(w, userOutboundData)
	if err != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "Login: EncoderError:", err)
		return
	}
}

func deathFilters(r *http.Request) (models.DeathFilter, error) {
	filtersCheck := models.DeathFilter{}

	var err error
	id := r.URL.Query().Get("gramPanchayatID")
	if id != "" {
		id = id[1 : len(id)-1]
		split := strings.Split(id, ",")
		filterIds := make([]int, 0)
		for i, _ := range split {
			splitId, err := strconv.Atoi(split[i])
			if err != nil {
				logrus.Print(err)
				return filtersCheck, err
			}
			filterIds = append(filterIds, splitId)
		}
		filtersCheck.GramPanchayatID = filterIds

	}

	id = r.URL.Query().Get("gaonId")
	if id != "" {
		id = id[1 : len(id)-1]
		split := strings.Split(id, ",")
		filterIds := make([]int, 0)
		for i, _ := range split {
			splitId, err := strconv.Atoi(split[i])
			if err != nil {
				logrus.Print(err)
				return filtersCheck, err
			}
			filterIds = append(filterIds, splitId)
		}
		filtersCheck.GaonID = filterIds

	}

	id = r.URL.Query().Get("taskID")
	if id != "" {
		id = id[1 : len(id)-1]
		split := strings.Split(id, ",")
		filterIds := make([]int, 0)
		for i, _ := range split {
			splitId, err := strconv.Atoi(split[i])
			if err != nil {
				logrus.Print(err)
				return filtersCheck, err
			}
			filterIds = append(filterIds, splitId)
		}
		filtersCheck.TaskID = filterIds

	}

	id = r.URL.Query().Get("tehsilID")
	if id != "" {
		id = id[1 : len(id)-1]
		split := strings.Split(id, ",")
		filterIds := make([]int, 0)
		for i, _ := range split {
			splitId, err := strconv.Atoi(split[i])
			if err != nil {
				logrus.Print(err)
				return filtersCheck, err
			}
			filterIds = append(filterIds, splitId)
		}
		filtersCheck.TehsilID = filterIds

	}

	id = r.URL.Query().Get("blockID")
	if id != "" {
		id = id[1 : len(id)-1]
		split := strings.Split(id, ",")
		filterIds := make([]int, 0)
		for i, _ := range split {
			splitId, err := strconv.Atoi(split[i])
			if err != nil {
				logrus.Print(err)
				return filtersCheck, err
			}
			filterIds = append(filterIds, splitId)
		}
		filtersCheck.BlockID = filterIds

	}

	tasks := r.URL.Query().Get("taskName")
	tasks = strings.Replace(tasks, "'", "", -1)
	if tasks != "" {
		tasks = tasks[1 : len(tasks)-1]
		split := strings.Split(tasks, ",")
		filtersCheck.TaskName = split
	}

	fromDate := r.URL.Query().Get("fromDate")
	if fromDate != "" {
		var timeErr error
		filtersCheck.FromDate, timeErr = time.Parse("02-01-2006", fromDate)
		if timeErr != nil {
			logrus.Print(timeErr)
			return filtersCheck, timeErr
		}
	}

	toDate := r.URL.Query().Get("toDate")
	if toDate != "" {
		var timeErr error
		filtersCheck.ToDate, timeErr = time.Parse("02-01-2006", toDate)
		if timeErr != nil {
			logrus.Print(timeErr)
			return filtersCheck, timeErr
		}
	}

	filtersCheck.Search = r.URL.Query().Get("search")

	filtersCheck.Status = r.URL.Query().Get("status")
	sort := r.URL.Query().Get("sort")
	if sort == "asc" {
		filtersCheck.IsAscending = true
	}

	filtersCheck.OrderBy = r.URL.Query().Get("orderBy")

	var limit int
	var page int
	strLimit := r.URL.Query().Get("limit")
	if strLimit == "" {
		limit = defaultLimit
	} else {
		limit, err = strconv.Atoi(strLimit)
		if err != nil {
			logrus.Printf("Limit: cannot get limit:%v", err)
			return filtersCheck, err
		}
	}

	strPage := r.URL.Query().Get("page")
	if strPage == "" {
		page = 0
	} else {
		page, err = strconv.Atoi(strPage)
		if err != nil {
			logrus.Printf("Page: cannot get page:%v", err)
			return filtersCheck, err
		}
	}

	filtersCheck.Limit = limit
	filtersCheck.Page = page

	return filtersCheck, nil
}
func filters(r *http.Request) (models.FiltersCheck, error) {
	filtersCheck := models.FiltersCheck{}
	isSearched := false
	searchedName := r.URL.Query().Get("name")
	if searchedName != "" {
		isSearched = true
	}

	var limit int
	var err error
	var page int
	strLimit := r.URL.Query().Get("limit")
	if strLimit == "" {
		limit = defaultLimit
	} else {
		limit, err = strconv.Atoi(strLimit)
		if err != nil {
			logrus.Printf("Limit: cannot get limit:%v", err)
			return filtersCheck, err
		}
	}

	strPage := r.URL.Query().Get("page")
	if strPage == "" {
		page = 0
	} else {
		page, err = strconv.Atoi(strPage)
		if err != nil {
			logrus.Printf("Page: cannot get page:%v", err)
			return filtersCheck, err
		}
	}

	filtersCheck = models.FiltersCheck{
		IsSearched:   isSearched,
		SearchedName: searchedName,
		Page:         page,
		Limit:        limit}
	return filtersCheck, nil
}
func AddRole(w http.ResponseWriter, r *http.Request) {
	var roleDetails models.RoleDetails

	decoderErr := utilities.Decoder(r, &roleDetails)
	if decoderErr != nil {
		utilities.HandlerError(w, http.StatusBadRequest, "AddRole: Decoder error:", decoderErr)
		return
	}

	err := helper.AddRole(roleDetails)
	if err != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "AddRole: cannot add role:", err)
		return
	}
}
func BulkAddRole(w http.ResponseWriter, r *http.Request) {
	roleDetails := make([]models.RoleDetails, 0)

	decoderErr := utilities.Decoder(r, &roleDetails)
	if decoderErr != nil {
		utilities.HandlerError(w, http.StatusBadRequest, "AddRole: Decoder error:", decoderErr)
		return
	}

	roleIds, err := helper.BulkAddRole(roleDetails)
	if err != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "AddRole: cannot add role:", err)
		return
	}
	encErr := utilities.Encoder(w, roleIds)
	if encErr != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "Failed to encode :", err)
		return
	}

}

func AddGaon(w http.ResponseWriter, r *http.Request) {
	var userDetails models.GaonDetails
	decoderErr := utilities.Decoder(r, &userDetails)
	if decoderErr != nil {
		utilities.HandlerError(w, http.StatusBadRequest, "AddGaon: Decoder error:", decoderErr)
		return
	}

	// transaction started
	txErr := database.Tx(func(tx *sqlx.Tx) error {
		userAndRoleID, err := helper.GetUserByPhoneNo(userDetails.LekhPalPhone, tx)
		if err != nil {
			return err
		}

		userID, err := helper.AddUser(userDetails.LekhPalName, userDetails.LekhPalPhone, utilities.LekhPal, userAndRoleID, tx)
		if err != nil {
			return err
		}

		gaonID, err := helper.AddGaon(userDetails, tx)
		if err != nil {
			return err
		}

		err = helper.AddUserGaon(userID, gaonID, tx)
		return err
	})
	if txErr != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "AddGaon: transaction error:", txErr)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func AddSdm(w http.ResponseWriter, r *http.Request) {
	var userDetails models.TehsilCreateRequest
	decoderErr := utilities.Decoder(r, &userDetails)
	if decoderErr != nil {
		utilities.HandlerError(w, http.StatusBadRequest, "AddSdm: Decoder error:", decoderErr)
		return
	}

	// transaction started
	txErr := database.Tx(func(tx *sqlx.Tx) error {
		userAndRoleID, err := helper.GetUserByPhoneNo(userDetails.PhoneNo, tx)
		if err != nil {
			return err
		}

		userID, err := helper.AddUser(userDetails.Name, userDetails.PhoneNo, utilities.SDM, userAndRoleID, tx)
		if err != nil {
			return err
		}

		tehsilID, err := helper.AddTehsil(userDetails.Tehsil, tx)
		if err != nil {
			return err
		}

		err = helper.AddUserTehsil(userID, tehsilID, tx)
		return err
	})
	if txErr != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "AddSdm: transaction error:", txErr)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func AddGramPanchayatInformation(w http.ResponseWriter, r *http.Request) {
	var userDetails models.GramPanchayatCreateRequest
	decoderErr := utilities.Decoder(r, &userDetails)
	if decoderErr != nil {
		utilities.HandlerError(w, http.StatusBadRequest, "AddGramPanchayatInformation: Decoder error:%v", decoderErr)
		return
	}
	if userDetails.SachivPhoneNo == userDetails.SahayakPhone {
		utilities.HandlerError(w, http.StatusBadRequest, "Sahayak and Sachiv cannot have same phone no.", errors.New("sahayak and sachiv same phone no"))
		return
	}
	txErr := database.Tx(func(tx *sqlx.Tx) error {
		//TODO: We can have multiple gram panchayat with same Sachiv. Please check for that.
		userAndRoleID, err := helper.GetUserByPhoneNo(userDetails.SachivPhoneNo, tx)
		if err != nil {
			return err
		}

		sachivID, err := helper.AddUser(userDetails.SachivName, userDetails.SachivPhoneNo, utilities.Sachiv, userAndRoleID, tx)
		if err != nil {
			return err
		}

		userAndRoleID, err = helper.GetUserByPhoneNo(userDetails.SahayakPhone, tx)
		if err != nil {
			return err
		}

		sahayakID, err := helper.AddUser(userDetails.SahayakName, userDetails.SahayakPhone, utilities.Sahayak, userAndRoleID, tx)
		if err != nil {
			return err
		}

		gramPanchayatID, err := helper.AddGramPanchayat(userDetails.GramPanchayat, userDetails.TehsilID, userDetails.BlockID, tx)
		if err != nil {
			return err
		}

		err = helper.AddUserGramPanchayat(sachivID, gramPanchayatID, tx)
		if err != nil {
			return err
		}
		err = helper.AddUserGramPanchayat(sahayakID, gramPanchayatID, tx)
		return err
	})
	if txErr != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "AddGramPanchayatInformation: AddGramPanchayat:", txErr)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func GetUserInfo(w http.ResponseWriter, r *http.Request) {
	contextValues, ok := r.Context().Value(utilities.UserContextKey).(models.ContextValues)
	if !ok {
		utilities.HandlerError(w, http.StatusInternalServerError, "GetUserInfo: Context for details:", errors.New("cannot get context details"))
		return
	}
	info, err := helper.GetUserInfo(contextValues.ID)
	if err != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "GetUserInfo:GetGramPanchayatInformation: cannot get list of gram panchayat:", err)
		return
	}

	err = utilities.Encoder(w, &info)
	if err != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "GetUserInfo: GetGramPanchayatInformation: EncoderError:", err)
		return
	}
}

func GetAdminInfo(w http.ResponseWriter, r *http.Request) {
	contextValues, ok := r.Context().Value(utilities.UserContextKey).(models.ContextValues)
	if !ok {
		utilities.HandlerError(w, http.StatusInternalServerError, "GetUserInfo: Context for details:", errors.New("cannot get context details"))
		return
	}
	info, err := helper.GetAdminInfo(contextValues.ID)
	if err != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "GetUserInfo:GetGramPanchayatInformation: cannot get list of gram panchayat:", err)
		return
	}

	err = utilities.Encoder(w, &info)
	if err != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "GetUserInfo: GetGramPanchayatInformation: EncoderError:", err)
		return
	}
}

func GetGramPanchayatInformation(w http.ResponseWriter, r *http.Request) {
	filterCheck, err := filters(r)
	if err != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "GetGramPanchayatInformation: cannot get filters properly: ", err)
		return
	}

	listOfGramPanchayat, err := helper.GetGramPanchayatList(&filterCheck)
	if err != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "GetGramPanchayatInformation: cannot get list of gram panchayat:", err)
		return
	}

	err = utilities.Encoder(w, &listOfGramPanchayat)
	if err != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "GetGramPanchayatInformation: EncoderError:", err)
		return
	}
}

func GetTehsilList(resp http.ResponseWriter, req *http.Request) {
	limit, page, _, err := utilities.GetLimitAndPage(req)
	if err != nil {
		utilities.HandlerError(resp, http.StatusInternalServerError, "GetTehsilList: Failed to get limit and page", err)
		return
	}

	TehsilList, err := helper.GetAllTehsil(limit, page)
	if err != nil {
		utilities.HandlerError(resp, http.StatusInternalServerError, "GetTehsilList: Failed to get tehsil names", err)
		return
	}

	encErr := utilities.Encoder(resp, TehsilList)
	if encErr != nil {
		utilities.HandlerError(resp, http.StatusInternalServerError, "GetTehsilList: Failed to encode output", encErr)
		return
	}

}

//func GetDeathDetails(w http.ResponseWriter, r *http.Request) {
//	filterCheck, err := filters(r)
//	if err != nil {
//		utilities.HandlerError(w, http.StatusInternalServerError, "GetDeathDetails: filters: cannot get filters:%v", err)
//	}
//	deathDuration := r.URL.Query().Get("deathDuration")
//	DeathDetails, err := helper.GetDeathDetails(deathDuration, filterCheck)
//	if err != nil {
//		utilities.HandlerError(w, http.StatusInternalServerError, "GetDeathDetails: cannot get death details:%v", err)
//		return
//	}
//}

func GetTehsils(resp http.ResponseWriter, req *http.Request) {
	//limit, page, _, err := utilities.GetLimitAndPage(req)
	//if err != nil {
	//	utilities.HandlerError(resp, http.StatusInternalServerError, "GetTehsils: Failed to get filter for tehsils", err)
	//	return
	//}

	filterCheck, err := filters(req)
	if err != nil {
		utilities.HandlerError(resp, http.StatusInternalServerError, "GetGramPanchayatInformation: cannot get filters properly: ", err)
		return
	}

	TehsilList, err := helper.GetTehsils(filterCheck)
	if err != nil {
		utilities.HandlerError(resp, http.StatusInternalServerError, "GetTehsils: Failed to get tehsils", err)
		return
	}

	encErr := utilities.Encoder(resp, TehsilList)
	if encErr != nil {
		utilities.HandlerError(resp, http.StatusInternalServerError, "GetTehsils: Failed to encode output", encErr)
		return
	}
}

func GetTasks(resp http.ResponseWriter, req *http.Request) {
	limit, page, searchText, err := utilities.GetLimitAndPage(req)
	if err != nil {
		utilities.HandlerError(resp, http.StatusInternalServerError, "GetTasks: Failed to get filter for tasks", err)
		return
	}

	tasks, err := helper.GetTasks(searchText, limit, page)
	if err != nil {
		utilities.HandlerError(resp, http.StatusInternalServerError, "GetTasks: Failed to get tasks", err)
		return
	}

	encErr := utilities.Encoder(resp, tasks)
	if encErr != nil {
		utilities.HandlerError(resp, http.StatusInternalServerError, "GetTasks: Failed to encode output", encErr)
		return
	}
}

func GetTotalDeaths(w http.ResponseWriter, _ *http.Request) {
	date := time.Now()
	month := date.Month()
	year, week := date.ISOWeek()
	TotalDeaths, err := helper.GetDeathCount(month, year, week)
	if err != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "GetTotalDeaths: Failed to get total no of deaths for month, week and today", err)
		return
	}

	encErr := utilities.Encoder(w, TotalDeaths)
	if encErr != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "GetTotalDeaths: Failed to encode the output", encErr)
		return
	}
}

func GetDistrictPostInfo(w http.ResponseWriter, _ *http.Request) {
	DistrictPostInfo, err := helper.GetDistrictPost()
	if err != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "GetDistrictPostInfo: Failed to fetch district level post info", err)
		return
	}
	districtPostInfoOutput := make([]models.DistrictPostOutput, 0)
	for i := range DistrictPostInfo {
		var out []models.TaskName
		err = json.Unmarshal(DistrictPostInfo[i].TaskName, &out)
		if err != nil {
			utilities.HandlerError(w, http.StatusInternalServerError, "GetDeaths: UnMarshal", err)
			return
		}
		output := models.DistrictPostOutput{
			Role:     DistrictPostInfo[i].Role,
			Name:     DistrictPostInfo[i].Name,
			PhoneNo:  DistrictPostInfo[i].PhoneNo,
			TaskName: out,
		}

		districtPostInfoOutput = append(districtPostInfoOutput, output)
	}
	encErr := utilities.Encoder(w, districtPostInfoOutput)
	if encErr != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "GetDistrictPostInfo: Failed to encode the output", encErr)
		return
	}
}

func GetGraph(w http.ResponseWriter, r *http.Request) {
	deathFilters, err := deathFilters(r)
	if err != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "cannot get death filters properly", err)
		return
	}

	graphDetails, err := helper.GetGraphTesting(deathFilters)
	if err != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "GetGraph: cannot get graph details:", err)
		return
	}

	encErr := utilities.Encoder(w, graphDetails)
	if encErr != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "GetGraph: Failed to encode the output", encErr)
		return
	}
}

func GetDeathDetailsAdmin(w http.ResponseWriter, r *http.Request) {
	deathFilters, err := deathFilters(r)
	if err != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "cannot get death filters properly", err)
		return
	}

	deathDetails, err := helper.GetDeathsAdmin(deathFilters)
	if err != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "cannot get death details", err)
		return
	}
	deathDetailsOutput := make([]models.DeathDetailsOutput, 0)
	for i := range deathDetails {
		var out []models.TaskDetail
		err = json.Unmarshal(deathDetails[i].TaskDetails, &out)
		if err != nil {
			utilities.HandlerError(w, http.StatusInternalServerError, "GetDeaths: UnMarshal", err)
			return
		}
		var registeredByDetails models.UserInfo
		if deathDetails[i].CreatedBy != 0 {
			registeredByDetails, err = helper.GetUserInfo(deathDetails[i].CreatedBy)
			if err != nil {
				utilities.HandlerError(w, http.StatusInternalServerError, "FetchDeathReview: unable to get reviewer info", err)
				return
			}
		}

		deathDetailsOut := models.DeathDetailsOutput{
			ID:                deathDetails[i].ID,
			Name:              deathDetails[i].Name,
			PhoneNo:           deathDetails[i].PhoneNo,
			Age:               deathDetails[i].Age,
			Gender:            deathDetails[i].Gender,
			AadharNumber:      deathDetails[i].AadharNumber,
			Status:            deathDetails[i].Status,
			Address:           deathDetails[i].Address,
			CreatedBy:         deathDetails[i].CreatedBy,
			RegisterBy:        registeredByDetails,
			CreatedAt:         deathDetails[i].CreatedAt,
			DateOfDeath:       deathDetails[i].DateOfDeath,
			GramPanchayatId:   deathDetails[i].GramPanchayatId,
			GramPanchayatName: deathDetails[i].GramPanchayatName,
			GaonId:            deathDetails[i].GaonId,
			GaonName:          deathDetails[i].GaonName,
			TehsilId:          deathDetails[i].TehsilId,
			TehsilName:        deathDetails[i].TehsilName,
			BlockId:           deathDetails[i].BlockId,
			BlockName:         deathDetails[i].BlockName,
			TaskDetails:       out,
		}
		deathDetailsOutput = append(deathDetailsOutput, deathDetailsOut)

	}

	err = utilities.Encoder(w, deathDetailsOutput)
	if err != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "GetDeaths: EncoderError", err)
	}
}

func EditGramPanchayat(w http.ResponseWriter, r *http.Request) {
	var gramPanchayatDetails models.GramUserDetails

	decoderErr := utilities.Decoder(r, &gramPanchayatDetails)
	if decoderErr != nil {
		utilities.HandlerError(w, http.StatusBadRequest, "EditGramPanchayat: Decoder error:", decoderErr)
		return
	}

	txErr := database.Tx(func(tx *sqlx.Tx) error {
		err := helper.EditGramPanchayat(gramPanchayatDetails, tx)
		if err != nil {
			utilities.HandlerError(w, http.StatusInternalServerError, "EditGramPanchayat: cannot edit gramPanchayat:", err)
			return err
		}

		err = helper.EditUser(gramPanchayatDetails.SachivName, gramPanchayatDetails.SachivPhoneNo, gramPanchayatDetails.SachivID, tx)
		if err != nil {
			return err
		}

		err = helper.EditUser(gramPanchayatDetails.SahayakName, gramPanchayatDetails.SahayakPhoneNo, gramPanchayatDetails.SahayakID, tx)
		if err != nil {
			return err
		}

		return nil
	})
	if txErr != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "AddGramPanchayatInformation: AddGramPanchayat:", txErr)
		return
	}
}

func EditTehsil(w http.ResponseWriter, r *http.Request) {
	var tehsilDetails models.TehsilUserDetails

	decoderErr := utilities.Decoder(r, &tehsilDetails)
	if decoderErr != nil {
		utilities.HandlerError(w, http.StatusBadRequest, "AddSdm: Decoder error:", decoderErr)
		return
	}

	txErr := database.Tx(func(tx *sqlx.Tx) error {
		err := helper.EditTehsil(tehsilDetails, tx)
		if err != nil {
			utilities.HandlerError(w, http.StatusInternalServerError, "EditGramPanchayat: cannot edit gramPanchayat:", err)
			return err
		}

		err = helper.EditUser(tehsilDetails.Name, tehsilDetails.PhoneNo, tehsilDetails.UserID, tx)
		return err
	})
	if txErr != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "AddGramPanchayatInformation: AddGramPanchayat:", txErr)
		return
	}
}

func GetBlock(w http.ResponseWriter, _ *http.Request) {
	blockDetails, err := helper.GetBlockDetails()
	if err != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "GetBlock: Failed to fetch blockDetails details", err)
		return
	}
	encErr := utilities.Encoder(w, blockDetails)
	if encErr != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "GetBlock: Failed to encode output", encErr)
		return
	}
}

func AddBlock(w http.ResponseWriter, r *http.Request) {
	var blockDetail models.BlockDetails
	decoderErr := utilities.Decoder(r, &blockDetail)
	if decoderErr != nil {
		utilities.HandlerError(w, http.StatusBadRequest, "AddBlock: Decoder error:", decoderErr)
		return
	}

	id, err := helper.AddBlock(blockDetail)
	if err != nil {
		utilities.HandlerError(w, http.StatusBadRequest, "AddBlock: Failed to add block:", decoderErr)
		return
	}
	encErr := utilities.Encoder(w, id)
	if encErr != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "GetBlock: Failed to encode output", encErr)
		return
	}
}

func EditBlock(w http.ResponseWriter, r *http.Request) {
	var blockDetail models.BlockDetails
	decoderErr := utilities.Decoder(r, &blockDetail)
	if decoderErr != nil {
		utilities.HandlerError(w, http.StatusBadRequest, "EditBlock: Decoder error:", decoderErr)
		return
	}

	err := helper.UpdateBlock(blockDetail)
	if err != nil {
		utilities.HandlerError(w, http.StatusBadRequest, "EditBlock: Failed to update block:", decoderErr)
		return
	}
}

func FetchDeathReview(w http.ResponseWriter, r *http.Request) {
	deathFilters, err := deathFilters(r)
	if err != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "cannot get death filters properly", err)
		return
	}

	deathDetails, err := helper.GetDeathReview(deathFilters)
	if err != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "FetchDeathReview: failed to get death review details", err)
		return
	}
	deathDetailsOutput := make([]models.DeathDetailsOutput, 0)
	for i := range deathDetails {
		out := make([]models.TaskDetail, 0)
		err = json.Unmarshal(deathDetails[i].TaskDetails, &out)
		if err != nil {
			utilities.HandlerError(w, http.StatusInternalServerError, "FetchDeathReview: UnMarshal", err)
			return
		}
		var reviewerDetail models.UserInfo
		if deathDetails[i].ReviewedBy.Int64 != 0 {
			reviewerDetail, err = helper.GetUserInfo(int(deathDetails[i].ReviewedBy.Int64))
			if err != nil {
				utilities.HandlerError(w, http.StatusInternalServerError, "FetchDeathReview: unable to get reviewer info", err)
				return
			}
		}
		var registeredByDetails models.UserInfo
		if deathDetails[i].CreatedBy != 0 {
			registeredByDetails, err = helper.GetUserInfo(deathDetails[i].CreatedBy)
			if err != nil {
				utilities.HandlerError(w, http.StatusInternalServerError, "FetchDeathReview: unable to get reviewer info", err)
				return
			}
		}
		deathDetailsOut := models.DeathDetailsOutput{
			ID:                deathDetails[i].ID,
			DeathId:           deathDetails[i].DeathId,
			Name:              deathDetails[i].Name,
			PhoneNo:           deathDetails[i].PhoneNo,
			Age:               deathDetails[i].Age,
			Gender:            deathDetails[i].Gender,
			AadharNumber:      deathDetails[i].AadharNumber,
			Status:            deathDetails[i].Status,
			Address:           deathDetails[i].Address,
			CreatedBy:         deathDetails[i].CreatedBy,
			CreatedAt:         deathDetails[i].CreatedAt,
			DateOfDeath:       deathDetails[i].DateOfDeath,
			GramPanchayatId:   deathDetails[i].GramPanchayatId,
			GramPanchayatName: deathDetails[i].GramPanchayatName,
			GaonId:            deathDetails[i].GaonId,
			GaonName:          deathDetails[i].GaonName,
			TehsilId:          deathDetails[i].TehsilId,
			TehsilName:        deathDetails[i].TehsilName,
			BlockId:           deathDetails[i].BlockId,
			BlockName:         deathDetails[i].BlockName,
			TaskDetails:       out,
			IsReviewed:        deathDetails[i].IsReviewed,
			Comment:           deathDetails[i].Comment,
			ReviewedAt:        deathDetails[i].ReviewedAt,
			ReviewedBy:        deathDetails[i].ReviewedBy,
			Reviewer:          reviewerDetail,
			RegisterBy:        registeredByDetails,
		}
		deathDetailsOutput = append(deathDetailsOutput, deathDetailsOut)

	}
	encErr := utilities.Encoder(w, deathDetailsOutput)
	if encErr != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "FetchDeathReview: Failed to encode output", encErr)
		return
	}
}

func ReviewDeathDetails(w http.ResponseWriter, r *http.Request) {
	var deathDetailsReview models.RandomDeath

	decoderErr := utilities.Decoder(r, &deathDetailsReview)
	if decoderErr != nil {
		utilities.HandlerError(w, http.StatusBadRequest, "ReviewDeathDetails: Decoder error:", decoderErr)
		return
	}

	if deathDetailsReview.ReviewComment == "" {
		utilities.HandlerError(w, http.StatusBadRequest, "review cannot be empty", errors.New("review comment cannot be empty"))
		return
	}

	contextValues, ok := r.Context().Value(utilities.UserContextKey).(models.ContextValues)
	if !ok {
		utilities.HandlerError(w, http.StatusInternalServerError, "DeathDetails: Context for details:", errors.New("cannot get context details"))
		return
	}

	err := helper.ReviewDeathDetails(deathDetailsReview, contextValues.ID)
	if err != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "ReviewDeathDetails: cannot review death", err)
		return
	}
}

func GetGaon(w http.ResponseWriter, r *http.Request) {
	GaonDetails, err := helper.GetGaon()
	if err != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "GetGaon: cannot get list of gram panchayat:", err)
		return
	}

	err = utilities.Encoder(w, &GaonDetails)
	if err != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "GetGaon: EncoderError:", err)
		return
	}

}

func EditGaon(w http.ResponseWriter, r *http.Request) {
	var gaonDetail models.GaonDetails
	decoderErr := utilities.Decoder(r, &gaonDetail)
	if decoderErr != nil {
		utilities.HandlerError(w, http.StatusBadRequest, "EditGaon: Decoder error:", decoderErr)
		return
	}

	txErr := database.Tx(func(tx *sqlx.Tx) error {
		err := helper.UpdateGaon(gaonDetail)
		if err != nil {
			utilities.HandlerError(w, http.StatusBadRequest, "EditGaon: Failed to update gaon:", err)
			return err
		}

		err = helper.EditUser(gaonDetail.LekhPalName, gaonDetail.LekhPalPhone, gaonDetail.LekhPalID, tx)
		return err
	})
	if txErr != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "AddGramPanchayatInformation: AddGramPanchayat:", txErr)
		return
	}
}
