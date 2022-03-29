package httprequest

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
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
			return fmt.Errorf("unable to unmarshal json body: %v", err)
		}
	case MIMEApplicationXml, MIMETextXml:
		err = xml.Unmarshal(respBytes, &resp)
		if err != nil {
			return fmt.Errorf("unable to unmarshal xml body: %v", err)
		}
	default:
		return fmt.Errorf("unsupported content type: %s", b.contentType)
	}

	return nil
}

func (b *RequestBuilder) Build(ctx context.Context) (*http.Request, error) {
	var body []byte
	var err error

	body, err = b.resolveContentType()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, b.httpMethod, b.url, bytes.NewReader(body))
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

func (b *RequestBuilder) resolveContentType() (body []byte, err error) {
	b.SetHeader(HeaderContentType, b.contentType)

	// Parse the content type using mime parsing and save the mediatype as the content type
	// parameters are currently unsupported and are discarded
	b.contentType, _, err = mime.ParseMediaType(b.contentType)
	if err != nil {
		return nil, fmt.Errorf("unable to parse media type: %v", err)
	}

	switch b.contentType {
	case MIMEApplicationJson:
		body, err = json.Marshal(b.body)
		if err != nil {
			return nil, fmt.Errorf("unable to marshal body to json: %v", err)
		}
	case MIMEApplicationXml, MIMETextXml:
		body, err = xml.Marshal(b.body)
		if err != nil {
			return nil, fmt.Errorf("unable to marshal body to xml: %v", err)
		}
	default:
		return nil, fmt.Errorf("unsupported content type: %s", b.contentType)
	}

	return body, nil
}

type Doer interface {
	Do(*http.Request) (*http.Response, error)
}
