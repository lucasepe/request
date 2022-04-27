package request

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/lucasepe/request/body"
	"github.com/lucasepe/request/reply"
	"github.com/lucasepe/request/uritemplates"
)

func NewPathParam(key string, val any) PathParam {
	s, err := toStringE(val)
	if err == nil {
		return PathParam{Name: key, Value: err.Error()}
	}

	return PathParam{Name: key, Value: s}
}

type PathParam struct {
	Name  string
	Value string
}

type Request struct {
	verb         string
	baseUrl      string
	path         string
	pathParams   map[string]string
	headers      map[string]string
	bodyGetter   body.Getter
	handler      reply.Handler
	client       *http.Client
	roundTripper http.RoundTripper
}

func Get(baseUrl string) *Request {
	return newRequest("GET", baseUrl)
}

func Head(baseUrl string) *Request {
	return newRequest("HEAD", baseUrl)
}

func Post(baseUrl string) *Request {
	return newRequest("POST", baseUrl)
}

func Put(baseUrl string) *Request {
	return newRequest("PUT", baseUrl)
}

func Patch(baseUrl string) *Request {
	return newRequest("PATCH", baseUrl)
}

func Delete(baseUrl string) *Request {
	return newRequest("DELETE", baseUrl)
}

func (r *Request) Path(path string, params ...PathParam) *Request {
	r.path = path
	for _, el := range params {
		r.pathParams[el.Name] = el.Value
	}
	return r
}

// Header sets a header on a request. It overwrites the existing values of a key.
func (r *Request) Header(key string, value string) *Request {
	r.headers[key] = value
	return r
}

// Client sets the http.Client to use for requests. If nil, it uses http.DefaultClient.
func (r *Request) Client(cl *http.Client) *Request {
	r.client = cl
	return r
}

// Transport sets the http.RoundTripper to use for requests.
// If set, it makes a shallow copy of the http.Client before modifying it.
func (r *Request) Transport(rt http.RoundTripper) *Request {
	r.roundTripper = rt
	return r
}

// Body sets the BodyGetter to use to build the body of a request.
// The provided BodyGetter is used as an http.Request.GetBody func.
// It implicitly sets method to POST.
func (r *Request) Body(src body.Getter) *Request {
	r.bodyGetter = src
	return r
}

// Into decodes a response as a JSON object.
func (r *Request) Into(v any) *Request {
	r.handler = reply.ToAny(v)
	return r
}

// IntoString writes the response body to the provided string pointer.
func (r *Request) IntoString(sp *string) *Request {
	r.handler = reply.ToString(sp)
	return r
}

// IntoBufioReader takes a callback which wraps the response body in a bufio.Reader.
func (r *Request) IntoBufioReader(f func(r *bufio.Reader) error) *Request {
	r.handler = reply.ToBufioReader(f)
	return r
}

// IntoBytesBuffer writes the response body to the provided bytes.Buffer.
func (r *Request) IntoBytesBuffer(buf *bytes.Buffer) *Request {
	r.handler = reply.ToBytesBuffer(buf)
	return r
}

// IntoWriter copies the response body to w.
func (r *Request) IntoWriter(w io.Writer) *Request {
	r.handler = reply.ToWriter(w)
	return r
}

// ReplyHandler specify the response handler.
func (r *Request) ReplyHandler(h reply.Handler) *Request {
	r.handler = h
	return r
}

func (r *Request) Do(ctx context.Context) error {
	req, err := r.build(ctx)
	if err != nil {
		return err
	}

	cl := DefaultClient()
	if r.client != nil {
		cl = r.client
	}

	if r.roundTripper != nil {
		cl2 := *cl
		cl2.Transport = r.roundTripper
		cl = &cl2
	}

	res, err := cl.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	handler := consumeBody
	if r.handler != nil {
		handler = r.handler
	}

	return handler(res)
}

func (r *Request) build(ctx context.Context) (*http.Request, error) {
	u, err := buildUrl(r.baseUrl, r.path, r.pathParams)
	if err != nil {
		return nil, err
	}

	var body io.ReadCloser
	if r.bodyGetter != nil {
		body, err = r.bodyGetter()
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, r.verb, u.String(), body)
	if err != nil {
		return nil, err
	}

	req.GetBody = r.bodyGetter

	for k, v := range r.headers {
		req.Header.Set(k, v)
	}

	return req, nil
}

func buildUrl(baseUrl, path string, vars map[string]string) (*url.URL, error) {
	u, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}

	_, unescaped, err := uritemplates.Expand(path, vars)
	if err != nil {
		return nil, err
	}

	ref, err := url.Parse(unescaped)
	if err != nil {
		return nil, err
	}

	return u.ResolveReference(ref), nil
}

func newRequest(verb string, baseUrl string) *Request {
	return &Request{
		verb:       verb,
		baseUrl:    baseUrl,
		pathParams: make(map[string]string),
		headers:    make(map[string]string),
	}
}

func consumeBody(res *http.Response) (err error) {
	const maxDiscardSize = 640 * 1 << 10
	if _, err = io.CopyN(io.Discard, res.Body, maxDiscardSize); err == io.EOF {
		err = nil
	}
	return err
}
