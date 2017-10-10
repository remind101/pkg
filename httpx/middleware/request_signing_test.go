package middleware

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	httpsignatures "github.com/99designs/httpsignatures-go"
	"github.com/remind101/pkg/httpx"
	"golang.org/x/net/context"
)

type fakeHandler struct {
}

func (h *fakeHandler) ServeHTTPContext(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	print(w, "ok")
	return nil
}

func wrap(h httpx.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := h.ServeHTTPContext(context.Background(), w, r)
		switch err.(type) {
		case nil:
			w.WriteHeader(200)
		case RequestSignatureError:
			w.WriteHeader(403)
			print(w, err.Error())
		default:
			w.WriteHeader(500)
			print(w, err.Error())
		}
	}
}

func mustReadString(t *testing.T, body io.ReadCloser) string {
	result, err := ioutil.ReadAll(body)
	if err != nil {
		t.Fatal(err)
	}
	return string(result)
}

func mustGET(t *testing.T, h http.Handler, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest("GET", "http://example.com"+path, nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}

func mustGETSigned(t *testing.T, h http.Handler, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest("GET", "http://example.com"+path, nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	h.ServeHTTP(w, signTestRequest(req))
	return w
}

func signTestRequest(r *http.Request) *http.Request {
	httpsignatures.DefaultSha256Signer.SignRequest("test-key", "signing-key", r)
	return r
}

func TestRequestSigning(t *testing.T) {
	cfg := RequestSigningConfig{
		ForceVerification:    true,
		SigningKeyRepository: NewStaticSigningKeyRepository(map[string]string{"test-key": "signing-key"}),
	}
	h := VerifySignature(cfg, &fakeHandler{})

	r := mustGET(t, wrap(h), "http://example.com")

	if r.Code != 403 {
		t.Errorf("expected a signature verification error, got %s", mustReadString(t, r.Result().Body))
	}

	r = mustGETSigned(t, wrap(h), "http://example.com")

	if r.Code != 200 {
		t.Errorf("expected a no signature verification error, got %s", mustReadString(t, r.Result().Body))
	}
}

func TestOptionalRequestSigning(t *testing.T) {
	cfg := RequestSigningConfig{
		SigningKeyRepository: NewStaticSigningKeyRepository(map[string]string{"other-test-key": "signing-key"}),
	}
	h := VerifySignature(cfg, &fakeHandler{})

	r := mustGET(t, wrap(h), "http://example.com")

	if r.Code != 200 {
		t.Errorf("expected signature to be optional, got error %s", mustReadString(t, r.Result().Body))
	}

	r = mustGETSigned(t, wrap(h), "http://example.com")

	if r.Code != 200 {
		t.Errorf("expected missing signing key not to trigger an error, got %s", mustReadString(t, r.Result().Body))
	}
}
