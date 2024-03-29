package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/magiconair/properties/assert"
	"github.com/robertolopezlopez/immudemo/authentication"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
)

type (
	DBaseMock struct {
		mock.Mock
	}
	buffer struct {
		bytes.Buffer
		io.ReaderFrom // conflicts with and hides bytes.Buffer's ReaderFrom.
		io.WriterTo   // conflicts with and hides bytes.Buffer's WriterTo.
	}
)

func (m *DBaseMock) CreateTables(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

func (m *DBaseMock) Log(ctx context.Context, strings []string) error {
	return m.Called(ctx, strings).Error(0)
}

func (m *DBaseMock) Find(ctx context.Context, s string) ([]string, error) {
	args := m.Called(ctx, s)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), nil
}

func (m *DBaseMock) Count(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Get(0).(int), args.Error(1)
}

func setUpMock(m DBaseMock) *gin.Engine {
	router := setupRouter()
	dataBase = &m
	return router
}

func setUpRecorder(r *gin.Engine, method, url, requestBody string, authenticate bool) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	rb := new(buffer)
	_, err := rb.WriteString(requestBody)
	if err != nil {
		panic(err.Error())
	}
	req, _ := http.NewRequest(method, url, rb)
	if authenticate {
		req.Header.Add(authentication.AuthTokenHeader, authentication.AuthTokenValue)
	}
	r.ServeHTTP(w, req)
	return w
}

func TestCountRoute_OK(t *testing.T) {
	m := DBaseMock{}
	m.On("Count", mock.Anything).Return(1, nil)
	router := setUpMock(m)

	w := setUpRecorder(router, http.MethodGet, "/count", "", true)

	assert.Equal(t, w.Code, http.StatusOK)
	assert.Equal(t, w.Body.String(), `{"count":1}`)

	m.AssertExpectations(t)
}

func TestCountRoute_auth_NOK(t *testing.T) {
	m := DBaseMock{}
	router := setUpMock(m)

	w := setUpRecorder(router, http.MethodGet, "/count", "", false)

	assert.Equal(t, w.Code, http.StatusUnauthorized)

	m.AssertExpectations(t)
}

func TestCountRoute_NOK(t *testing.T) {
	m := DBaseMock{}
	m.On("Count", mock.Anything).Return(0, fmt.Errorf("an error"))
	router := setUpMock(m)

	w := setUpRecorder(router, http.MethodGet, "/count", "", true)

	assert.Equal(t, w.Code, http.StatusInternalServerError)
	assert.Equal(t, w.Body.String(), `"an error"`)

	m.AssertExpectations(t)
}

func TestReadRoute_1_OK(t *testing.T) {
	m := DBaseMock{}
	m.On("Find", mock.Anything, "1").Return([]string{"hola"}, nil)
	router := setUpMock(m)

	w := setUpRecorder(router, http.MethodGet, "/?n=1", "", true)

	assert.Equal(t, w.Code, http.StatusOK)
	assert.Equal(t, w.Body.String(), `{"logs":["hola"]}`)

	m.AssertExpectations(t)
}

func TestReadRoute_all_OK(t *testing.T) {
	m := DBaseMock{}
	m.On("Find", mock.Anything, "").Return([]string{"hola", "Roberto", "adiós"}, nil)
	router := setUpMock(m)

	w := setUpRecorder(router, http.MethodGet, "/", "", true)

	assert.Equal(t, w.Code, http.StatusOK)
	assert.Equal(t, w.Body.String(), `{"logs":["hola","Roberto","adiós"]}`)

	m.AssertExpectations(t)
}

func TestReadRoute_all_NOK(t *testing.T) {
	m := DBaseMock{}
	m.On("Find", mock.Anything, "").Return([]string{}, fmt.Errorf("error db"))
	router := setUpMock(m)

	w := setUpRecorder(router, http.MethodGet, "/", "", true)

	assert.Equal(t, w.Code, http.StatusInternalServerError)
	assert.Equal(t, w.Body.String(), `"error db"`)

	m.AssertExpectations(t)
}

func TestWrite_OK(t *testing.T) {
	m := DBaseMock{}
	m.On("Log", mock.Anything, []string{"hola"}).Return(nil)
	router := setUpMock(m)

	w := setUpRecorder(router, http.MethodPost, "/", `{"msgs":["hola"]}`, true)

	assert.Equal(t, w.Code, http.StatusCreated)
	assert.Equal(t, w.Body.String(), `{"message":"message(s) successfully logged"}`)

	m.AssertExpectations(t)
}

func TestWrite_Log_NOK(t *testing.T) {
	m := DBaseMock{}
	m.On("Log", mock.Anything, []string{"hola"}).Return(fmt.Errorf("log error"))
	router := setUpMock(m)

	w := setUpRecorder(router, http.MethodPost, "/", `{"msgs":["hola"]}`, true)

	assert.Equal(t, w.Code, http.StatusInternalServerError)
	assert.Equal(t, w.Body.String(), `{"error":"log error"}`)

	m.AssertExpectations(t)
}

func TestWrite_no_body_NOK(t *testing.T) {
	m := DBaseMock{}
	router := setUpMock(m)

	w := setUpRecorder(router, http.MethodPost, "/", `{"msgs":[]}`, true)

	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Equal(t, w.Body.String(), `{"error":"msgs should not be empty"}`)

	m.AssertExpectations(t)
}
