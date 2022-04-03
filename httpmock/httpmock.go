package httpmock

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"

	"github.com/stretchr/testify/mock"
)

const (
	headerKeyContentType = "Content-Type"

	mimeApplicationJson = "application/json"
)

func NewMock() *Mock {
	return &Mock{}
}

type Mock struct {
	mock.Mock
}

func (m *Mock) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)

}

func (m *Mock) GET(url string) *HttpCall {
	header := http.Header{}
	header.Add(headerKeyContentType, mimeApplicationJson)
	matchOn := MatchOn{
		HttpMethod: http.MethodGet,
		Url:        url,
		Header:     header,
		Body:       nil,
	}

	requestMatcher := mock.MatchedBy(makeRequestMatcherFunc(matchOn))
	return &HttpCall{m.On("Do", requestMatcher)}
}

func (m *Mock) POST(url string, body interface{}) *HttpCall {
	header := http.Header{}
	header.Add(headerKeyContentType, mimeApplicationJson)
	matchOn := MatchOn{
		HttpMethod: http.MethodPost,
		Url:        url,
		Header:     header,
		Body:       body,
	}

	requestMatcher := mock.MatchedBy(makeRequestMatcherFunc(matchOn))
	return &HttpCall{m.On("Do", requestMatcher)}
}

type HttpCall struct {
	*mock.Call
}

func (c *HttpCall) Run(run func(req *http.Request)) *HttpCall {
	c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*http.Request))
	})
	return c
}

func (c *HttpCall) Return(statusCode int, out interface{}, outErr error) *HttpCall {
	// TODO Support multiple content types
	data, err := json.Marshal(out)
	if err != nil {
		panic(err)
	}

	resp := &http.Response{
		Status:        http.StatusText(statusCode),
		StatusCode:    statusCode,
		Body:          io.NopCloser(bytes.NewReader(data)),
		ContentLength: int64(len(data)),
	}

	c.Call.Return(resp, outErr)
	return c
}

func (c *HttpCall) AddHeader(key, val string) *HttpCall {
	//TODO
	return c
}

type MatchOn struct {
	HttpMethod string
	Url        string
	Header     http.Header
	Body       interface{}
}

func makeRequestMatcherFunc(matchOn MatchOn) func(*http.Request) bool {
	return func(request *http.Request) bool {
		if matchOn.HttpMethod != request.Method {
			return false
		}

		// Need to compare the request header to the matchOn header by
		// 1. Asserting all keys in the request header are in the matchOn header.
		// 2. Asserting that the length of the values for the key is the same
		// 3. Asserting that the values of the key are the same and in the same order
		for key, vals := range request.Header {
			strings.ToLower(key)
			wantVals, ok := matchOn.Header[key]
			if !ok {
				return false
			}
			if len(vals) != len(wantVals) {
				return false
			}
			for i, val := range vals {
				if val != wantVals[i] {
					return false
				}
			}
		}

		return checkBodyMatch(request, matchOn.Body)
	}
}

func checkBodyMatch(request *http.Request, wantBody interface{}) bool {
	if (request.Body == nil || request.Body == http.NoBody) && wantBody == nil {
		return true
	}

	if reflect.ValueOf(wantBody).Kind() == reflect.Ptr {
		panic("expected non-pointer type for body match")
	}

	reqBodyBytes, err := ioutil.ReadAll(request.Body)
	if err != nil {
		panic(err)
	}
	// Once the body is read, we need to reset the request body in case further reads on the body are made
	err = request.Body.Close()
	if err != nil {
		panic(err)
	}
	request.Body = io.NopCloser(bytes.NewReader(reqBodyBytes))

	// Unmarshal the body as a json.RawMessage and then marshal it again to ensure that order of keys does not
	// affect the equality check
	var body json.RawMessage
	err = json.Unmarshal(reqBodyBytes, &body)
	if err != nil {
		panic(err)
	}

	actualBodyBytes, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}

	expectedBodyBytes, err := json.Marshal(wantBody)
	if err != nil {
		panic(err)
	}

	return bytes.Compare(expectedBodyBytes, actualBodyBytes) == 0
}
