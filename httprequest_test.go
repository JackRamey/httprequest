package httprequest

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/jackramey/httprequest/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testUrl = "https://example.com/api/v1/endpoint"

type UserRequest struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	IsAdmin bool   `json:"isAdmin"`
}

type UserResponse struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	IsAdmin bool   `json:"isAdmin"`
}

var req1 = UserRequest{
	ID:      6,
	Name:    "jack",
	IsAdmin: true,
}

var resp1 = UserResponse{
	ID:      42,
	Name:    "stephen",
	IsAdmin: false,
}

func TestRequestBuilder_Build(t *testing.T) {
	t.Run("Builder with all fields returns a request", func(t *testing.T) {
		expectedBytes, err := json.Marshal(req1)
		require.NoError(t, err)

		req, err := New(http.MethodGet, testUrl, req1).
			Build(context.Background())
		require.NoError(t, err)
		assert.NotEmpty(t, req)
		assert.Equal(t, http.MethodGet, req.Method)
		assert.Equal(t, MIMEApplicationJson, req.Header.Get(HeaderContentType))

		bodyBytes, err := ioutil.ReadAll(req.Body)
		require.NoError(t, err)
		assert.Equal(t, expectedBytes, bodyBytes)
	})
	t.Run("Builder with added header is reflected in request", func(t *testing.T) {
		expectedBytes, err := json.Marshal(req1)
		require.NoError(t, err)

		req, err := New(http.MethodGet, testUrl, req1).
			AddHeader("foo", "bar").
			Build(context.Background())
		require.NoError(t, err)
		assert.NotEmpty(t, req)
		assert.Equal(t, http.MethodGet, req.Method)
		assert.Equal(t, MIMEApplicationJson, req.Header.Get(HeaderContentType))
		assert.Equal(t, "bar", req.Header.Get("foo"))

		bodyBytes, err := ioutil.ReadAll(req.Body)
		require.NoError(t, err)
		assert.Equal(t, expectedBytes, bodyBytes)
	})
	t.Run("Adding a content type is reflected in the request", func(t *testing.T) {
		expectedBytes, err := xml.Marshal(req1)
		require.NoError(t, err)

		req, err := New(http.MethodGet, testUrl, req1).
			ContentType(MIMEApplicationXml).
			Build(context.Background())
		require.NoError(t, err)
		assert.NotEmpty(t, req)
		assert.Equal(t, http.MethodGet, req.Method)
		assert.Equal(t, MIMEApplicationXml, req.Header.Get(HeaderContentType))

		bodyBytes, err := ioutil.ReadAll(req.Body)
		require.NoError(t, err)
		assert.Equal(t, expectedBytes, bodyBytes)
	})
	t.Run("Builder with invalid content type returns an error", func(t *testing.T) {
		_, err := New(http.MethodGet, testUrl, req1).
			ContentType("application/morse-code").
			Build(context.Background())
		require.Error(t, err)
	})
}

func TestRequestBuilder_validateStatusCode(t *testing.T) {
	tests := []struct {
		name                string
		expectedStatusCodes []int
		resp                *http.Response
		wantErr             assert.ErrorAssertionFunc
	}{
		{
			name:                "Response status code is in expected values returns no error",
			expectedStatusCodes: []int{http.StatusCreated, http.StatusAccepted},
			resp: &http.Response{
				StatusCode: http.StatusAccepted,
			},
			wantErr: assert.NoError,
		},
		{
			name:                "Response status code is http.StatusOK and expected values is nil",
			expectedStatusCodes: nil,
			resp: &http.Response{
				StatusCode: http.StatusOK,
			},
			wantErr: assert.NoError,
		},
		{
			name:                "Response status code is not in expected values returns an error",
			expectedStatusCodes: []int{http.StatusCreated, http.StatusAccepted},
			resp: &http.Response{
				StatusCode: http.StatusNotFound,
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &RequestBuilder{
				expectedStatusCodes: tt.expectedStatusCodes,
			}
			tt.wantErr(t, b.validateStatusCode(tt.resp), fmt.Sprintf("validateStatusCode(%v)", tt.resp))
		})
	}
}

func TestRequestBuilder_Do(t *testing.T) {
	t.Run("Do GET", func(t *testing.T) {
		ctx := context.Background()
		mock := httpmock.NewMock()
		mock.GET(testUrl).Return(http.StatusOK, resp1, nil)

		var out UserResponse
		resp, err := New(http.MethodGet, testUrl, nil).Do(ctx, mock, &out)
		require.NoError(t, err)
		require.NotEmpty(t, resp)
		require.NotEmpty(t, out)
		assert.Equal(t, resp1.ID, out.ID)
		assert.Equal(t, resp1.Name, out.Name)
		assert.Equal(t, resp1.IsAdmin, out.IsAdmin)
		mock.AssertExpectations(t)
	})
	t.Run("Do POST", func(t *testing.T) {
		ctx := context.Background()
		mock := httpmock.NewMock()
		mock.POST(testUrl, req1).Return(http.StatusOK, resp1, nil)

		var out UserResponse
		resp, err := New(http.MethodPost, testUrl, req1).Do(ctx, mock, &out)
		require.NoError(t, err)
		require.NotEmpty(t, resp)
		require.NotEmpty(t, out)
		assert.Equal(t, resp1.ID, out.ID)
		assert.Equal(t, resp1.Name, out.Name)
		assert.Equal(t, resp1.IsAdmin, out.IsAdmin)
		mock.AssertExpectations(t)
	})
}
