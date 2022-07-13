package service

import (
	"bytes"
	"coinbit-test/lib/handler"
	"coinbit-test/lib/helper"
	"coinbit-test/lib/proto"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

const (
	ThresholdAmount = 10000
	ThresholdTime   = 2 * time.Minute
)

type ResponseDetails struct {
	Balance        float32 `json:"balance"`
	AboveThreshold bool    `json:"above_threshold"`
}

func Details(h *handler.Handler, w http.ResponseWriter, r *http.Request) {
	urlQuery := r.URL.Query()

	paramWalletID := urlQuery.Get("wallet_id")
	if paramWalletID == "" {
		helper.ResponseError(r, w, http.StatusBadRequest, errors.New("parameter wallet_id invalid"))
		return
	}

	walletID, err := strconv.ParseInt(paramWalletID, 10, 64)
	if err != nil {
		helper.ResponseError(r, w, http.StatusBadRequest, err)
		return
	}

	b, err := h.Storage.Get(generateKey(walletID))
	if err != nil {
		helper.ResponseError(r, w, http.StatusInternalServerError, err)
		return
	}

	deposits := make([]proto.Deposit, 0, 10)
	var balance float32
	var aboveThreshold bool

	if b != nil {
		err = json.Unmarshal(b, &deposits)
		if err != nil {
			helper.ResponseError(r, w, http.StatusInternalServerError, err)
			return
		}

		var amountInWindowTime float32
		for _, deposit := range deposits {
			// check windows time
			if time.Since(deposit.Timestamp.AsTime()) < ThresholdTime {
				amountInWindowTime += deposit.Amount
			}
			balance += deposit.Amount
		}
		if amountInWindowTime > ThresholdAmount {
			aboveThreshold = true
		}
	}

	helper.ResponseObject(w, ResponseDetails{
		Balance:        balance,
		AboveThreshold: aboveThreshold,
	})
}

func Deposit(h *handler.Handler, w http.ResponseWriter, r *http.Request) {
	buf, err := helper.GetBufBody(r)
	if err != nil {
		helper.ResponseError(r, w, http.StatusInternalServerError, err)
		return
	}

	r.Body = ioutil.NopCloser(bytes.NewReader(buf.Bytes()))

	var payload proto.Deposit
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		helper.ResponseError(r, w, http.StatusInternalServerError, err)
		return
	}

	b, err := json.Marshal(payload)
	if err != nil {
		helper.ResponseError(r, w, http.StatusInternalServerError, err)
		return
	}

	_, err = h.Emitter.Emit(generateKey(payload.WalletId), b)
	if err != nil {
		helper.ResponseError(r, w, http.StatusInternalServerError, err)
		return
	}

	helper.ResponseObject(w, payload)
}

func generateKey(walletID int64) string {
	return fmt.Sprintf("wallet_id:%d", walletID)
}
