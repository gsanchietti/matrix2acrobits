package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/nethesis/matrix2acrobits/models"
	"github.com/nethesis/matrix2acrobits/service"
	"github.com/labstack/echo/v4"
)

const adminTokenHeader = "X-Super-Admin-Token"

// RegisterRoutes wires API endpoints to Echo handlers.
func RegisterRoutes(e *echo.Echo, svc *service.MessageService, adminToken string) {
	h := handler{svc: svc, adminToken: adminToken}
	e.POST("/api/client/send_message", h.sendMessage)
	e.POST("/api/client/fetch_messages", h.fetchMessages)
	e.POST("/api/internal/map_sms_to_matrix", h.postMapping)
	e.GET("/api/internal/map_sms_to_matrix", h.getMapping)
}

type handler struct {
	svc        *service.MessageService
	adminToken string
}

func (h handler) sendMessage(c echo.Context) error {
	var req models.SendMessageRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
	}

	resp, err := h.svc.SendMessage(c.Request().Context(), &req)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(http.StatusOK, resp)
}

func (h handler) fetchMessages(c echo.Context) error {
	var req models.FetchMessagesRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
	}

	resp, err := h.svc.FetchMessages(c.Request().Context(), &req)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(http.StatusOK, resp)
}

func (h handler) postMapping(c echo.Context) error {
	if err := h.ensureAdminAccess(c); err != nil {
		return err
	}

	var req models.MappingRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
	}
	resp, err := h.svc.SaveMapping(&req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, resp)
}

func (h handler) getMapping(c echo.Context) error {
	if err := h.ensureAdminAccess(c); err != nil {
		return err
	}

	smsNumber := strings.TrimSpace(c.QueryParam("sms_number"))
	if smsNumber == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "sms_number query parameter is required")
	}

	resp, err := h.svc.LookupMapping(smsNumber)
	if err != nil {
		if errors.Is(err, service.ErrMappingNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, resp)
}

func (h handler) ensureAdminAccess(c echo.Context) error {
	if h.adminToken == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "admin token not configured")
	}
	if !isLocalhost(c.RealIP()) {
		return echo.NewHTTPError(http.StatusForbidden, "mapping API only available from localhost")
	}
	token := c.Request().Header.Get(adminTokenHeader)
	if token == "" || token != h.adminToken {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid admin token")
	}
	return nil
}

func mapServiceError(err error) error {
	switch {
	case errors.Is(err, service.ErrAuthentication):
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	case errors.Is(err, service.ErrInvalidRecipient):
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrMappingNotFound):
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
}

func isLocalhost(ip string) bool {
	trimmed := ip
	if colon := strings.LastIndex(trimmed, ":"); colon != -1 {
		trimmed = trimmed[:colon]
	}
	switch trimmed {
	case "127.0.0.1", "::1", "localhost":
		return true
	default:
		return false
	}
}
