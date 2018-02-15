package middleware

import (
	"fmt"
	"net/http"
	"strings"

	httpsignatures "github.com/99designs/httpsignatures-go"
	"github.com/pkg/errors"
	"github.com/remind101/pkg/httpx"
	"context"
)

// RequestSignatureError should be used in your error handling middleware.
// Usage:
//
//   import "github.com/tomasen/realip"
//   import "github.com/remind101/pkg/logger"
//   import "github.com/remind101/pkg/metrics"
//   import "github.com/remind101/pkg/httpx/middleware"
//   ...
//   switch err := errors.Cause(err).(type) {
//     case middleware.RequestSignatureError:
//       remoteAddr := realip.RealIP(r)
//       metrics.Count("authentication.failure", 1, map[string]string{"keyid": err.KeyID, "remote_ip": remoteAddr}, 1.0)
//       logger.Error(ctx, "authentication failure", "keyid", err.KeyID, "remote_ip", remoteAddr, "err", err.Error())
//       w.WriteHeader(403)
//       fmt.Fprintf(w, `{"error":"request signature verification error"}`)
//     ...
//   }
//   ...
type RequestSignatureError struct {
	KeyID string
	msg   string
}

func newRequestSignatureError(keyID, msg string) RequestSignatureError {
	return RequestSignatureError{
		KeyID: keyID,
		msg:   msg,
	}
}

func (e RequestSignatureError) Error() string {
	return fmt.Sprintf(`Request signature error: keyID=%s: %s`, e.KeyID, e.msg)
}

// VerifySignature wraps an httpx.Handler with a request signature check.
// Usage:
//
//   import "github.com/remind101/pkg/httpx"
//   import "github.com/remind101/pkg/httpx/middleware"
//   ...
//   r := httpx.NewRouter()
//   keys := middleware.NewStaticSigningKeyRepositoryFromStringSlice([]string{"key_id:key_secret", "key2_id:key2_secret"})
//   cfg := middleware.RequestSigningConfig{ForceVerification: true, SigningKeyRepository: keys}
//   r.Handle("/foo", VerifySignature(cfg, myHandler)).Methods("GET")
//   ...
//
// See also documentation for RequestSigningConfig
// See https://tools.ietf.org/html/draft-cavage-http-signatures-07 for more details.
func VerifySignature(cfg RequestSigningConfig, h httpx.Handler) httpx.HandlerFunc {
	pass := h.ServeHTTPContext

	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		// net/http parses the Host header and puts it into r.Host, but we may be using it to calculate the signature
		if r.Header.Get("Host") == "" {
			r.Header.Add("Host", r.Host)
		}

		sig, err := httpsignatures.FromRequest(r)

		if err != nil {
			if cfg.ForceVerification {
				return newRequestSignatureError("", err.Error())
			}
			fmt.Printf("Skipping request verification: malformed/absent signature header: %s\n", err.Error())
			return pass(ctx, w, r)
		}

		key, err := cfg.GetKey(sig.KeyID)
		if err != nil {
			if cfg.ForceVerification {
				return errors.WithStack(err)
			}
			fmt.Printf("Skipping request verification: request signing key not found for keyID=%s\n", sig.KeyID)
			return pass(ctx, w, r)
		}

		if !sig.IsValid(key, r) {
			return newRequestSignatureError(sig.KeyID, "Bad request signature")
		}

		return pass(ctx, w, r)
	}
}

// RequestSigningConfig contains configuration for request signing middleware.
// ForceVerification - when true, rejects all requests with absent/malformed/invalid request signature header;
//                     when false, allows requests with absent/malformed request signature header, rejects
//                       requests with invalid signature.
// SigningKeyRepository - an implementation of SigningKeyRepository.
type RequestSigningConfig struct {
	ForceVerification bool
	SigningKeyRepository
}

// SigningKeyRepository stores request signing keys.
type SigningKeyRepository interface {
	GetKey(keyID string) (string, error)
}

// StaticSigningKeyRepository implements SigningKeyRepository and stores pairs of key ids
// and secrets in memory.
type StaticSigningKeyRepository struct {
	keys map[string]string
}

// NewStaticSigningKeyRepository creates a SigningKeyRepository from a string map where keys are
// key ids and values are HMAC secrets.
func NewStaticSigningKeyRepository(keys map[string]string) *StaticSigningKeyRepository {
	return &StaticSigningKeyRepository{keys}
}

// NewStaticSigningKeyRepositoryFromStringSlice creates a SigningKeyRepository from a string slice
// in form of []string{"keyId:keyValue"}, which can be used with StringSlice from https://github.com/urfave/cli.
func NewStaticSigningKeyRepositoryFromStringSlice(idkeys []string) *StaticSigningKeyRepository {
	keys := make(map[string]string, len(idkeys))
	for _, idkey := range idkeys {
		idAndKey := strings.Split(idkey, ":")
		if len(idAndKey) != 2 {
			panic("Bad request signing keys string!")
		}
		keys[idAndKey[0]] = idAndKey[1]
	}
	return NewStaticSigningKeyRepository(keys)
}

func (r *StaticSigningKeyRepository) GetKey(keyID string) (string, error) {
	key, ok := r.keys[keyID]
	if !ok {
		return "", newRequestSignatureError(keyID, "key not found")
	}
	return key, nil
}
