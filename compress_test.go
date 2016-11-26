package compress

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-http-utils/headers"
	"github.com/stretchr/testify/suite"
)

func init() {
	gw := gzip.NewWriter(&gziped)
	gw.Write(testContent)
	gw.Close()

	dw, _ := flate.NewWriter(&defalted, flate.DefaultCompression)
	dw.Write(testContent)
	dw.Close()
}

type CompressSuite struct {
	suite.Suite

	server *httptest.Server
}

var testContent = []byte("Hello，世界")
var gziped bytes.Buffer
var defalted bytes.Buffer

func (s *CompressSuite) SetupTest() {
	mux := http.NewServeMux()
	mux.Handle("/", Handler(http.HandlerFunc(helloHandlerFunc)))

	s.server = httptest.NewServer(mux)
}

func (s CompressSuite) TestCompressGzip() {
	defer s.server.Close()

	req, err := http.NewRequest(http.MethodGet, s.server.URL+"/", nil)
	s.Nil(err)

	req.Header.Set(headers.AcceptEncoding, "gzip")
	req.Header.Set(headers.ContentType, "text/plain")

	res, err := sendRequest(req)

	s.Nil(err)
	s.Equal(http.StatusOK, res.StatusCode)
	s.Equal(headers.AcceptEncoding, res.Header.Get(headers.Vary))
	s.Equal("gzip", res.Header.Get(headers.ContentEncoding))

	body, err := getResRawBody(res)
	s.Nil(err)
	s.Equal(gziped.Bytes(), body)
}

func (s CompressSuite) TestCompressDeflate() {
	defer s.server.Close()

	req, err := http.NewRequest(http.MethodGet, s.server.URL+"/", nil)
	s.Nil(err)

	req.Header.Set(headers.AcceptEncoding, "deflate")
	req.Header.Set(headers.ContentType, "text/plain")

	res, err := sendRequest(req)

	s.Nil(err)
	s.Equal(http.StatusOK, res.StatusCode)
	s.Equal(headers.AcceptEncoding, res.Header.Get(headers.Vary))
	s.Equal("deflate", res.Header.Get(headers.ContentEncoding))

	body, err := getResRawBody(res)
	s.Nil(err)
	s.Equal(defalted.Bytes(), body)
}

func (s CompressSuite) TestCompressGzipNegotiated() {
	defer s.server.Close()

	req, err := http.NewRequest(http.MethodGet, s.server.URL+"/", nil)
	s.Nil(err)

	req.Header.Set(headers.AcceptEncoding, "gzip, deflate")
	req.Header.Set(headers.ContentType, "text/plain")

	res, err := sendRequest(req)

	s.Nil(err)
	s.Equal(http.StatusOK, res.StatusCode)
	s.Equal(headers.AcceptEncoding, res.Header.Get(headers.Vary))
	s.Equal("gzip", res.Header.Get(headers.ContentEncoding))

	body, err := getResRawBody(res)
	s.Nil(err)
	s.Equal(gziped.Bytes(), body)
}

func (s CompressSuite) TestHead() {
	defer s.server.Close()

	req, err := http.NewRequest(http.MethodHead, s.server.URL+"/", nil)
	s.Nil(err)

	req.Header.Set(headers.AcceptEncoding, "gzip, deflate")
	req.Header.Set(headers.ContentType, "text/plain")

	res, err := sendRequest(req)

	s.Nil(err)
	s.Equal(http.StatusOK, res.StatusCode)
	s.NotEqual(headers.AcceptEncoding, res.Header.Get(headers.Vary))
	s.NotEqual("gzip", res.Header.Get(headers.ContentEncoding))
}

func (s CompressSuite) TestNotCompressible() {
	defer s.server.Close()

	req, err := http.NewRequest(http.MethodGet, s.server.URL+"/", nil)
	s.Nil(err)

	req.Header.Set(headers.AcceptEncoding, "gzip")
	req.Header.Set(headers.ContentType, "image/png")

	res, err := sendRequest(req)

	s.Nil(err)
	s.Equal(http.StatusOK, res.StatusCode)
	s.NotEqual(headers.AcceptEncoding, res.Header.Get(headers.Vary))
	s.NotEqual("gzip", res.Header.Get(headers.ContentEncoding))
}

func (s CompressSuite) TestNotMatchCompressible() {
	defer s.server.Close()

	req, err := http.NewRequest(http.MethodGet, s.server.URL+"/", nil)
	s.Nil(err)

	req.Header.Set(headers.AcceptEncoding, "not-match")
	req.Header.Set(headers.ContentType, "text/html")

	res, err := sendRequest(req)

	s.Nil(err)
	s.Equal(http.StatusNotAcceptable, res.StatusCode)
	s.NotEqual(headers.AcceptEncoding, res.Header.Get(headers.Vary))

	body, err := getResRawBody(res)
	s.Nil(err)
	s.Equal([]byte("supported encodings: gzip, deflate"), body)
}

func (s CompressSuite) TestAlreadySetContentEncoding() {
	defer s.server.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set(headers.ContentEncoding, "gzip")
		Handler(http.HandlerFunc(helloHandlerFunc)).ServeHTTP(res, req)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	req, err := http.NewRequest(http.MethodGet, server.URL+"/", nil)
	s.Nil(err)

	req.Header.Set(headers.AcceptEncoding, "gzip")
	req.Header.Set(headers.ContentType, "image/png")

	res, err := sendRequest(req)

	s.Nil(err)
	s.Equal(http.StatusOK, res.StatusCode)
	s.NotEqual(headers.AcceptEncoding, res.Header.Get(headers.Vary))
	s.Equal("gzip", res.Header.Get(headers.ContentEncoding))
}

func TestCompress(t *testing.T) {
	suite.Run(t, new(CompressSuite))
}

func helloHandlerFunc(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(http.StatusOK)

	res.Write(testContent)
}

func noContentFunc(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(http.StatusNoContent)

	res.Write(nil)
}

func sendRequest(req *http.Request) (*http.Response, error) {
	cli := &http.Client{}
	return cli.Do(req)
}

func getResRawBody(res *http.Response) ([]byte, error) {
	return ioutil.ReadAll(res.Body)
}
