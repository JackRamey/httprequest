package httprequest

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testUrl = "https://example.com/api/v1/endpoint"

type UserRequest struct {
	ID      int    `json:"ID"`
	Name    string `json:"name"`
	IsAdmin bool   `json:"isAdmin"`
}

var req1 = UserRequest{
	ID:      6,
	Name:    "jack",
	IsAdmin: true,
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
