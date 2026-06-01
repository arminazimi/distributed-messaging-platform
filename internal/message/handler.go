package message

import (
	"encoding/json"
	"errors"
	"fmt"
	"gateway/app"
	"gateway/config"
	"gateway/internal/balance"
	"gateway/internal/model"
	"gateway/internal/outbox"
	"gateway/pkg/tracing"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// SendHandler godoc
// @Summary      Send message request
// @Description  Deducts balance, enqueues a message for processing, returns processing ack
// @Tags         message
// @Accept       json
// @Produce      json
// @Param        request body model.Message true "message request"
// @Success      200 {object} map[string]any "ack with message_identifier"
// @Failure      400 {string} string "invalid input"
// @Failure      402 {string} string "dont have Not Enough Balance"
// @Failure      500 {string} string "internal error"
// @Router       /messages/send [post]
func SendHandler(c echo.Context) error {
	var s model.Message
	if err := json.NewDecoder(c.Request().Body).Decode(&s); err != nil {
		app.Logger.Error("invalid input ", "err", err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid input")
	}

	ctxWithUser := tracing.WithUser(c.Request().Context(), fmt.Sprint(s.CustomerID))
	c.SetRequest(c.Request().WithContext(ctxWithUser))

	if len(s.Recipients) == 0 {
		app.Logger.Error("zero recipients")
		return echo.NewHTTPError(http.StatusBadRequest, "zero recipients")
	}

	s.MessageIdentifier = uuid.NewString()
	// Atomic: deduct balance (user_transactions) + insert outbox (pending) in ONE DB transaction.
	tx, err := app.DB.BeginTxx(c.Request().Context(), nil)
	if err != nil {
		app.Logger.Error("begin tx", "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}
	defer func() {
		_ = tx.Rollback()
	}()

	transactionID, err := balance.ChargeTx(c.Request().Context(), tx, balance.ChargeRequest{
		CustomerID: s.CustomerID,
		Quantity:   len(s.Recipients),
		Type:       s.Type,
	})
	if err != nil {
		if errors.Is(err, balance.ErrInsufficientBalance) {
			app.Logger.Error("User Has Not Enough Balance ", "user id ", s.CustomerID)
			return echo.NewHTTPError(http.StatusPaymentRequired, "dont have Not Enough Balance ")
		}
		app.Logger.Error("ChargeTx ", "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}
	s.TransactionID = transactionID

	priority := 0
	if s.Type == model.EXPRESS {
		priority = 10
	}

	// Initial state: PENDING (inserted with the outbox record)
	if err := InsertPendingTx(c.Request().Context(), tx, s); err != nil {
		app.Logger.Error("insert message pending", "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}

	// Store the message in the outbox for the publisher to forward to RabbitMQ.
	if err := outbox.InsertTx(c.Request().Context(), tx, outbox.Event{
		AggregateType: "message",
		AggregateID:   s.MessageIdentifier,
		EventType:     "message.send",
		Priority:      priority,
		Status:        outbox.StatusPending,
		Payload: map[string]any{
			"exchange":       config.MessageExchange,
			"routing_key":    getQueue(s.Type),
			"message":        s,
			"transaction_id": s.TransactionID,
		},
	}); err != nil {
		app.Logger.Error("insert outbox", "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}

	if err := tx.Commit(); err != nil {
		app.Logger.Error("commit tx", "err", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status":             "processing",
		"message_identifier": s.MessageIdentifier,
	})
}

// HistoryHandler godoc
// @Summary      Get message history for user
// @Description  Returns sent message history for a user
// @Tags         message
// @Accept       json
// @Produce      json
// @Param        user_id query string true "User ID"
// @Param        status query string false "Filter by status (pending|sending|done|failed)"
// @Param        message_identifier query string false "Filter by message_identifier"
// @Success      200 {object} map[string]any
// @Failure      400 {string} string "user_id is required"
// @Failure      500 {string} string "internal error"
// @Router       /messages/history [get]
func HistoryHandler(c echo.Context) error {
	userID := c.QueryParam("user_id")
	if userID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "user_id is required")
	}

	status := c.QueryParam("status")
	messageIdentifier := c.QueryParam("message_identifier")

	history, err := GetUserHistory(c.Request().Context(), userID, status, messageIdentifier)
	if err != nil {
		app.Logger.Error("get message history", "user_id", userID, "status", status, "message_identifier", messageIdentifier, "err", err)
		return err
	}

	out := map[string]any{}
	out["history"] = history

	return c.JSON(http.StatusOK, out)
}

func getQueue(s model.Type) string {
	switch s {
	case model.NORMAL:
		return config.NormalQueue
	case model.EXPRESS:
		return config.ExpressQueue
	default:
		return config.NormalQueue
	}
}
