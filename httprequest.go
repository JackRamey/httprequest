package httprequest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	MIMEApplicationJson = "application/json"
)

func New(httpMethod, url string, body interface{}) *RequestBuilder {
	return &RequestBuilder{
		body:        body,
		url:         url,
		httpMethod:  httpMethod,
		contentType: MIMEApplicationJson,
	}
}

type RequestBuilder struct {
	body        interface{}
	url         string
	httpMethod  string
	contentType string
	header      http.Header
}

func (b *RequestBuilder) Do(ctx context.Context, client Doer, resp interface{}) error {
	req, err := b.Build(ctx)
	if err != nil {
		return err
	}

	_resp, err := client.Do(req)
	if err != nil {
		return err
	}

	respBytes, err := ioutil.ReadAll(_resp.Body)
	switch b.contentType {
	case MIMEApplicationJson:
		err = json.Unmarshal(respBytes, &resp)
		if err != nil {
			return errors.New("unable to unmarshal body to json")
		}
	default:
		return fmt.Errorf("unsupported content type: %s", b.contentType)
	}

	return nil
}

func (b *RequestBuilder) Build(ctx context.Context) (*http.Request, error) {
	var body []byte
	var err error

	switch b.contentType {
	case MIMEApplicationJson:
		body, err = json.Marshal(body)
		if err != nil {
			return nil, errors.New("unable to marshal body to json")
		}
	default:
		return nil, fmt.Errorf("unsupported content type: %s", b.contentType)
	}

	return http.NewRequestWithContext(ctx, b.httpMethod, b.url, bytes.NewReader(body))
}

func (b *RequestBuilder) ContentType(contentType string) *RequestBuilder {
	b.contentType = contentType
	return b
}

func (b *RequestBuilder) AddHeader(key, value string) *RequestBuilder {
	b.header.Add(key, value)
	return b
}

func (b *RequestBuilder) SetHeader(key, value string) *RequestBuilder {
	b.header.Set(key, value)
	return b
}

type Doer interface {
	Do(*http.Request) (*http.Response, error)
}
