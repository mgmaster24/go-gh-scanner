package main

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"
)

type jwksCache struct {
	uri      string
	issuer   string
	audience string

	mu        sync.RWMutex
	keys      map[string]*rsa.PublicKey
	fetchedAt time.Time
}

type jwksResponse struct {
	Keys []jwk `json:"keys"`
}

type jwk struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func newJWKSCache(uri, issuer, audience string) *jwksCache {
	return &jwksCache{uri: uri, issuer: issuer, audience: audience}
}

// validate extracts and validates the Bearer JWT from an Authorization header.
// Returns the token claims on success.
func (c *jwksCache) validate(authHeader string) (map[string]any, error) {
	if c.uri == "" {
		return nil, errors.New("JWKS_URI not configured")
	}

	token, ok := strings.CutPrefix(authHeader, "Bearer ")
	if !ok || token == "" {
		return nil, errors.New("missing bearer token")
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}

	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid token header: %w", err)
	}
	var header struct {
		Kid string `json:"kid"`
		Alg string `json:"alg"`
	}
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return nil, fmt.Errorf("malformed token header: %w", err)
	}
	if header.Alg != "RS256" {
		return nil, fmt.Errorf("unsupported algorithm: %s", header.Alg)
	}

	key, err := c.key(header.Kid)
	if err != nil {
		return nil, err
	}

	hash := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid token signature encoding: %w", err)
	}
	if err := rsa.VerifyPKCS1v15(key, crypto.SHA256, hash[:], sig); err != nil {
		return nil, errors.New("invalid signature")
	}

	claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid token claims encoding: %w", err)
	}
	var claims map[string]any
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, fmt.Errorf("malformed token claims: %w", err)
	}

	exp, _ := claims["exp"].(float64)
	if int64(exp) == 0 || time.Now().Unix() > int64(exp) {
		return nil, errors.New("token expired or missing exp")
	}

	if c.issuer != "" {
		if iss, _ := claims["iss"].(string); iss != c.issuer {
			return nil, fmt.Errorf("unexpected issuer: %s", iss)
		}
	}

	if c.audience != "" {
		if !hasAudience(claims["aud"], c.audience) {
			return nil, fmt.Errorf("audience %s not present in token", c.audience)
		}
	}

	return claims, nil
}

func hasAudience(aud any, want string) bool {
	switch v := aud.(type) {
	case string:
		return v == want
	case []any:
		for _, a := range v {
			if s, ok := a.(string); ok && s == want {
				return true
			}
		}
	}
	return false
}

// key returns the RSA public key for the given kid, refreshing JWKS when stale.
func (c *jwksCache) key(kid string) (*rsa.PublicKey, error) {
	const ttl = 24 * time.Hour

	c.mu.RLock()
	k, hit := c.keys[kid]
	stale := time.Since(c.fetchedAt) > ttl
	c.mu.RUnlock()

	if hit && !stale {
		return k, nil
	}

	if err := c.refresh(); err != nil {
		// Return stale key rather than failing if we have one.
		if hit {
			return k, nil
		}
		return nil, fmt.Errorf("JWKS fetch failed: %w", err)
	}

	c.mu.RLock()
	k = c.keys[kid]
	c.mu.RUnlock()
	if k == nil {
		return nil, fmt.Errorf("no key found for kid %q", kid)
	}
	return k, nil
}

func (c *jwksCache) refresh() error {
	resp, err := http.Get(c.uri) //nolint:noctx
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var j jwksResponse
	if err := json.Unmarshal(body, &j); err != nil {
		return err
	}

	keys := make(map[string]*rsa.PublicKey, len(j.Keys))
	for _, k := range j.Keys {
		if k.Kty != "RSA" {
			continue
		}
		pub, err := jwkToRSA(k)
		if err != nil {
			continue
		}
		keys[k.Kid] = pub
	}

	c.mu.Lock()
	c.keys = keys
	c.fetchedAt = time.Now()
	c.mu.Unlock()
	return nil
}

func jwkToRSA(k jwk) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
	if err != nil {
		return nil, err
	}
	n := new(big.Int).SetBytes(nBytes)
	e := int(new(big.Int).SetBytes(eBytes).Int64())
	if e == 0 {
		return nil, errors.New("invalid RSA exponent")
	}
	return &rsa.PublicKey{N: n, E: e}, nil
}
