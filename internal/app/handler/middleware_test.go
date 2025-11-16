package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestLoggingMiddleware(t *testing.T) {
	e := echo.New()
	mw := LoggingMiddleware(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := mw(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestMetricsMiddleware(t *testing.T) {
	e := echo.New()
	mw := MetricsMiddleware(func(c echo.Context) error {
		time.Sleep(10 * time.Millisecond)
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := mw(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}
