package httprequest

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
)

const (
	MIMEApplicationJson = "application/json"
	MIMEApplicationXml  = "application/xml"
	MIMETextXml         = "text/xml"

	HeaderAuthorization = "Authorization"
	HeaderContentType   = "Content-Type"
)

func New(httpMethod, url string, body interface{}) *RequestBuilder {
	return &RequestBuilder{
		body:                body,
		url:                 url,
		httpMethod:          httpMethod,
		expectedStatusCodes: []int{http.StatusOK},
		contentType:         MIMEApplicationJson,
	}
}

type RequestBuilder struct {
	body                interface{}
	url                 string
	httpMethod          string
	contentType         string
	expectedStatusCodes []int
	header              http.Header
}

func (b *RequestBuilder) Do(ctx context.Context, doer Doer, out interface{}) (*http.Response, error) {
	req, err := b.Build(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := doer.Do(req)
	if err != nil {
		return nil, err
	}

	err = b.validateStatusCode(resp)
	if err != nil {
		return nil, err
	}

	err = b.unmarshalResponse(resp, out)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (b *RequestBuilder) Build(ctx context.Context) (*http.Request, error) {
	var body io.Reader
	var err error

	body, err = b.resolveContentType()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, b.httpMethod, b.url, body)
	if err != nil {
		return nil, fmt.Errorf("unable to create request")
	}

	req.Header = b.header

	return req, nil
}

func (b *RequestBuilder) ContentType(contentType string) *RequestBuilder {
	b.contentType = contentType
	return b
}

func (b *RequestBuilder) StatusIs(status int) *RequestBuilder {
	b.expectedStatusCodes = []int{status}
	return b
}

func (b *RequestBuilder) StatusIn(statuses []int) *RequestBuilder {
	b.expectedStatusCodes = statuses
	return b
}

func (b *RequestBuilder) AddHeader(key, value string) *RequestBuilder {
	if b.header == nil {
		b.header = http.Header{}
	}

	b.header.Add(key, value)
	return b
}

func (b *RequestBuilder) SetHeader(key, value string) *RequestBuilder {
	if b.header == nil {
		b.header = http.Header{}
	}

	b.header.Set(key, value)
	return b
}

func (b *RequestBuilder) resolveContentType() (body io.Reader, err error) {
	if b.body == nil {
		return http.NoBody, nil
	}

	b.SetHeader(HeaderContentType, b.contentType)

	// Parse the content type using mime parsing and save the mediatype as the content type
	// parameters are currently unsupported and are discarded
	b.contentType, _, err = mime.ParseMediaType(b.contentType)
	if err != nil {
		return nil, fmt.Errorf("unable to parse media type: %v", err)
	}

	var bodyBytes []byte
	switch b.contentType {
	case MIMEApplicationJson:
		bodyBytes, err = json.Marshal(b.body)
		if err != nil {
			return nil, fmt.Errorf("unable to marshal body to json: %v", err)
		}
		body = bytes.NewReader(bodyBytes)
	case MIMEApplicationXml, MIMETextXml:
		bodyBytes, err = xml.Marshal(b.body)
		if err != nil {
			return nil, fmt.Errorf("unable to marshal body to xml: %v", err)
		}
		body = bytes.NewReader(bodyBytes)
	default:
		return nil, fmt.Errorf("unsupported content type: %s", b.contentType)
	}

	return body, nil
}

func (b *RequestBuilder) unmarshalResponse(resp *http.Response, out interface{}) error {
	respBytes, err := ioutil.ReadAll(resp.Body)
	switch b.contentType {
	case MIMEApplicationJson:
		err = json.Unmarshal(respBytes, &out)
		if err != nil {
			return fmt.Errorf("unable to unmarshal json body: %v", err)
		}
	case MIMEApplicationXml, MIMETextXml:
		err = xml.Unmarshal(respBytes, &out)
		if err != nil {
			return fmt.Errorf("unable to unmarshal xml body: %v", err)
		}
	default:
		return fmt.Errorf("unsupported content type: %s", b.contentType)
	}

	return nil
}

func (b *RequestBuilder) validateStatusCode(resp *http.Response) error {
	if len(b.expectedStatusCodes) == 0 {
		b.expectedStatusCodes = []int{http.StatusOK}
	}

	var isExpectedStatus bool
	for _, code := range b.expectedStatusCodes {
		if isExpectedStatus = resp.StatusCode == code; isExpectedStatus {
			break
		}
	}

	if !isExpectedStatus {
		return fmt.Errorf("received unexpected status code: %v", resp.StatusCode)
	}

	return nil
}

type Doer interface {
	Do(*http.Request) (*http.Response, error)
}
