package httpmock

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type InputData struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type OutputData struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Info      struct {
		Age int `json:"age"`
	} `json:"info"`
}

func TestNewMock(t *testing.T) {
	mock := NewMock()
	mock.GET("http://example.com").Return(http.StatusOK, OutputData{
		FirstName: "Jack",
		LastName:  "Ramey",
		Info: struct {
			Age int `json:"age"`
		}{Age: 34},
	}, nil)

	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	require.NoError(t, err)
	resp, err := mock.Do(req)
	require.NoError(t, err)
	assert.NotEmpty(t, resp)
	mock.AssertExpectations(t)
}
