package controllers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"iibb-imports/models"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"iibb-imports/infra"
)

func PostPayment(c *gin.Context) {
	w := c.Writer
	r := c.Request
	headerContentType := r.Header.Get("Content-Type")
	if headerContentType != "application/json" {
		errorResponse(w, "Content type is not application/json", http.StatusUnsupportedMediaType)
		return
	}
	var newPayment models.Payment
	var unmarshalErr *json.UnmarshalTypeError
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&newPayment)
	if err != nil {
		if errors.As(err, &unmarshalErr) {
			errorResponse(w, "Bad Request. Wrong type provided for field "+unmarshalErr.Field, http.StatusBadRequest)
		} else {
			errorResponse(w, "Bad Request "+err.Error(), http.StatusBadRequest)
		}
		return
	}
	ctx := context.Background()
	DBConection := infra.DbPayment
	tsql := fmt.Sprintf("USE [Interoperabilidad] INSERT INTO [dbo].[NotificationMOLPayment]([Key],[External_Reference],[Status],[Amount]) VALUES (@Key, @External_reference, @Status, @Amount);")
	result, err2 := DBConection.ExecContext(
		ctx,
		tsql,
		sql.Named("Key", newPayment.Key),
		sql.Named("External_reference", newPayment.External_reference),
		sql.Named("Status", newPayment.Status),
		sql.Named("Amount", newPayment.Rate),
	)
	if err2 != nil {
		errorResponse(w, "Error inserting new row: "+err2.Error(), http.StatusBadRequest)
		return
	} else if result != nil {
		errorResponse(w, "Successfully added new row", http.StatusCreated)
	}
}

func errorResponse(w http.ResponseWriter, message string, httpStatusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatusCode)
	resp := make(map[string]string)
	resp["message"] = message
	jsonResp, _ := json.Marshal(resp)
	w.Write(jsonResp)
}
