package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestMatrixAppTransaction(t *testing.T) {
	e := echo.New()

	// Minimal valid-ish transaction body (we don't parse it, only log)
	payload := `{"events":[], "other": "value"}`
	req := httptest.NewRequest(http.MethodPut, "/_matrix/app/v1/transactions/txn123", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetParamNames("txnId")
	c.SetParamValues("txn123")

	h := handler{}
	err := h.matrixAppTransaction(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}
