package handler

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"grampanchayat/database"
	"grampanchayat/database/helper"
	"grampanchayat/models"
	"grampanchayat/utilities"
	"net/http"
	"strconv"
)

func DeathRegistration(w http.ResponseWriter, r *http.Request) {
	var deathDetails models.DeathRegistrationRequest

	decoderErr := utilities.Decoder(r, &deathDetails)
	if decoderErr != nil {
		utilities.HandlerError(w, http.StatusBadRequest, "DeathRegistration: Decoder error:", decoderErr)
		return
	}

	contextValues, ok := r.Context().Value(utilities.UserContextKey).(models.ContextValues)
	if !ok {
		utilities.HandlerError(w, http.StatusInternalServerError, "DeathDetails: Context for details:", errors.New("cannot get context details"))
		return
	}

	// transaction started
	txErr := database.Tx(func(tx *sqlx.Tx) error {
		deathID, err := helper.DeathRegistration(deathDetails, contextValues.ID, tx)
		if err != nil {
			return err
		}

		deathDetails.ID = deathID

		addressId, err := helper.AddAddress(deathDetails.Address, tx)
		if err != nil {
			return err
		}

		err = helper.AddDeathAddress(deathID, addressId, tx)
		return err
	})
	if txErr != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "DeathRegistration ", txErr)
		return
	}
	userOutboundData := make(map[string]int)

	userOutboundData["Successfully Registered death: ID is"] = deathDetails.ID

	err := utilities.Encoder(w, userOutboundData)
	if err != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "DeathRegistration: EncoderError", err)
		return
	}
}

func GetDeathsNew(w http.ResponseWriter, r *http.Request) {
	GetDeaths(w, r, "new")
}

func GetDeathsProcessing(w http.ResponseWriter, r *http.Request) {
	GetDeaths(w, r, "processing")
}

func GetDeathsCompleted(w http.ResponseWriter, r *http.Request) {
	GetDeaths(w, r, "completed")
}

func GetDeaths(w http.ResponseWriter, r *http.Request, status string) {
	contextValues, ok := r.Context().Value(utilities.UserContextKey).(models.ContextValues)
	if !ok {
		utilities.HandlerError(w, http.StatusInternalServerError, "GetDeaths: Context for details:", errors.New("cannot get context details"))
		return
	}

	displayTaskTypes, _ := helper.GetDisplayTypes(contextValues.Role)
	actionableTaskTypes, _ := helper.GetActionableTaskTypes(contextValues.Role)
	search := r.URL.Query().Get("search")
	deathDetails, err := helper.GetDeathsNew(displayTaskTypes, status, contextValues.ID, search)
	if err != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "GetDeaths: cannot get new deaths", err)
		return
	}
	deathDetailsOutput := make([]models.DeathDetailsOutput, 0)
	for i, _ := range deathDetails {
		var out []models.TaskDetail
		err = json.Unmarshal(deathDetails[i].TaskDetails, &out)
		if err != nil {
			utilities.HandlerError(w, http.StatusInternalServerError, "GetDeaths: UnMarshal", err)
			return
		}
		var finalTaskDetails []models.TaskDetail
		for j, _ := range out {
			if funk.ContainsString(actionableTaskTypes, out[j].TaskType) {
				out[j].IsEditable = true
				finalTaskDetails = append(finalTaskDetails, out[j])
			} else if contextValues.Role == utilities.Sachiv || contextValues.Role == utilities.Sahayak {
				//non editable but show
				finalTaskDetails = append(finalTaskDetails, out[j])
			}
		}
		deathDetailsOut := models.DeathDetailsOutput{
			ID:           deathDetails[i].ID,
			Name:         deathDetails[i].Name,
			PhoneNo:      deathDetails[i].PhoneNo,
			Age:          deathDetails[i].Age,
			Gender:       deathDetails[i].Gender,
			AadharNumber: deathDetails[i].AadharNumber,
			Status:       deathDetails[i].Status,
			Address:      deathDetails[i].Address,
			CreatedBy:    deathDetails[i].CreatedBy,
			CreatedAt:    deathDetails[i].CreatedAt,
			DateOfDeath:  deathDetails[i].DateOfDeath,
			TaskDetails:  out,
		}
		deathDetailsOutput = append(deathDetailsOutput, deathDetailsOut)

	}

	err = utilities.Encoder(w, deathDetailsOutput)
	if err != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "GetDeaths: EncoderError", err)
		return
	}
}

func ProcessingTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.Atoi(chi.URLParam(r, "taskID"))
	if err != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "MarkProcessing: cannot get task id", err)
		return
	}

	var processing models.Processing
	err = utilities.Decoder(r, &processing)
	if err != nil {
		utilities.HandlerError(w, http.StatusBadRequest, "DeathRegistration: Decoder error:", err)
		return
	}

	if processing.Started {
		err = helper.MarkProcessing(taskID)
		if err != nil {
			utilities.HandlerError(w, http.StatusInternalServerError, "MarkProcessing: cannot update task as processing", err)
			return
		}
		message := "successfully marked task as processing"
		err = utilities.Encoder(w, &message)
		if err != nil {
			utilities.HandlerError(w, http.StatusInternalServerError, "MarkProcessing: EncoderError", err)
			return
		}
	} else {
		err := helper.TaskRejected(taskID, processing)
		if err != nil {
			utilities.HandlerError(w, http.StatusInternalServerError, "Failed to reject task", err)
			return
		}
		message := "successfully rejected task"
		err = utilities.Encoder(w, &message)
		if err != nil {
			utilities.HandlerError(w, http.StatusInternalServerError, "MarkProcessing: EncoderError", err)
			return
		}
	}
}

func MarkCompleted(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.Atoi(chi.URLParam(r, "taskID"))
	if err != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "MarkCompleted: cannot get task id", err)
		return
	}

	err = helper.MarkCompleted(taskID)
	if err != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "MarkProcessing: cannot update task as completed", err)
		return
	}

	message := "successfully marked task as completed"
	err = utilities.Encoder(w, &message)
	if err != nil {
		utilities.HandlerError(w, http.StatusInternalServerError, "MarkCompleted: EncoderError", err)
		return
	}
}
