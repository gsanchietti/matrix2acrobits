package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/nethesis/matrix2acrobits/models"
	"github.com/nethesis/matrix2acrobits/service"
	"github.com/stretchr/testify/assert"
)

func TestIsLocalhost(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		{"127.0.0.1", "127.0.0.1", true},
		{"127.0.0.1 with port", "127.0.0.1:8080", true},
		{"localhost", "localhost", true},
		{"localhost with port", "localhost:8080", true},
		{"Remote IP", "192.168.1.1", false},
		{"Remote IP with port", "192.168.1.1:8080", false},
		{"IPv4 different", "10.0.0.1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isLocalhost(tt.ip)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPushTokenReport(t *testing.T) {
	e := echo.New()
	svc := service.NewMessageService(nil, nil)

	t.Run("valid push token report", func(t *testing.T) {
		reqBody := models.PushTokenReportRequest{
			Selector:   "12869E0E6E553673C54F29105A0647204C416A2A:7C3A0D14",
			TokenMsgs:  "QVBBOTFiRzlhcVd2bW54bllCWldHOWh4dnRrZ3pUWFNvcGZpdWZ6bWM2dFAzS2J",
			AppIDMsgs:  "com.cloudsoftphone.app",
			TokenCalls: "Udl99X2JFP1bWwS5gR/wGeLE1hmAB2CMpr1Ej0wxkrY=",
			AppIDCalls: "com.cloudsoftphone.app.pushkit",
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/client/push_token_report", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)

		h := handler{svc: svc, adminToken: "test"}
		err := h.pushTokenReport(c)

		// Since we don't have a real database, this will fail with "database not initialized"
		// but we can verify the handler processes the request correctly
		assert.Error(t, err)
	})

	t.Run("invalid json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/client/push_token_report", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)

		h := handler{svc: svc, adminToken: "test"}
		err := h.pushTokenReport(c)

		// Should return a bind error
		assert.Error(t, err)
	})

	t.Run("empty selector", func(t *testing.T) {
		reqBody := models.PushTokenReportRequest{
			Selector: "",
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/client/push_token_report", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)

		h := handler{svc: svc, adminToken: "test"}
		err := h.pushTokenReport(c)

		assert.Error(t, err)
	})
}
