package utilities

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"net/http"
)

type Key string

const (
	UserContextKey Key = "values"
	Sachiv             = "Sachiv"
	Sahayak            = "Sahayak"
	SDM                = "SDM"
	LekhPal            = "Lekhpal"
)

func Decoder(r *http.Request, inter interface{}) error {
	err := json.NewDecoder(r.Body).Decode(&inter)
	if err != nil {
		logrus.Printf("decoderr error:%v", err)
		return err
	}
	return nil
}

func Encoder(w http.ResponseWriter, inter interface{}) error {
	err := json.NewEncoder(w).Encode(&inter)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Printf("encoder error:%v", err)
		return err
	}
	return nil
}

func HandlerError(w http.ResponseWriter, statusCode int, errorMessage string, error error) {
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(struct {
		MsgToUser string `json:"messageToUser"`
		DevInfo   string `json:"additionalInfoForDev"`
	}{
		MsgToUser: errorMessage,
		DevInfo:   error.Error(),
	})
	if err != nil {
		logrus.Printf("Write error: %v", err)
	}
	logrus.Printf(errorMessage, error)
	return
}

func GetLimitAndPage(req *http.Request) (int, int, string, error) {
	// var limit, page int
	// if req.URL.Query().Get("limit") != "" {
	//	limitS, err := strconv.Atoi(req.URL.Query().Get("limit"))
	//	if err != nil {
	//		return 0, 0, searchText, err
	//	}
	//	limit = limitS
	// } else {
	//	limit = 10000
	// }
	//
	// if req.URL.Query().Get("page") != "" {
	//	pageS, err := strconv.Atoi(req.URL.Query().Get("page"))
	//	if err != nil {
	//		return 0, 0, searchText, err
	//	}
	//	page = pageS
	// } else {
	//	page = 0
	// }
	//
	searchText := req.URL.Query().Get("searchText")
	// return limit, page, searchText, nil
	return 10000, 0, searchText, nil
}
