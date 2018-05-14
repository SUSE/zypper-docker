// Copyright 2017 The etcd Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package auth

import (
	"context"
	"crypto/rsa"
	"io/ioutil"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"go.uber.org/zap"
)

type tokenJWT struct {
	lg         *zap.Logger
	signMethod string
	signKey    *rsa.PrivateKey
	verifyKey  *rsa.PublicKey
	ttl        time.Duration
}

func (t *tokenJWT) enable()                         {}
func (t *tokenJWT) disable()                        {}
func (t *tokenJWT) invalidateUser(string)           {}
func (t *tokenJWT) genTokenPrefix() (string, error) { return "", nil }

func (t *tokenJWT) info(ctx context.Context, token string, rev uint64) (*AuthInfo, bool) {
	// rev isn't used in JWT, it is only used in simple token
	var (
		username string
		revision uint64
	)

	parsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return t.verifyKey, nil
	})

	switch err.(type) {
	case nil:
		if !parsed.Valid {
			if t.lg != nil {
				t.lg.Warn("invalid JWT token", zap.String("token", token))
			} else {
				plog.Warningf("invalid jwt token: %s", token)
			}
			return nil, false
		}

		claims := parsed.Claims.(jwt.MapClaims)

		username = claims["username"].(string)
		revision = uint64(claims["revision"].(float64))
	default:
		if t.lg != nil {
			t.lg.Warn(
				"failed to parse a JWT token",
				zap.String("token", token),
				zap.Error(err),
			)
		} else {
			plog.Warningf("failed to parse jwt token: %s", err)
		}
		return nil, false
	}

	return &AuthInfo{Username: username, Revision: revision}, true
}

func (t *tokenJWT) assign(ctx context.Context, username string, revision uint64) (string, error) {
	// Future work: let a jwt token include permission information would be useful for
	// permission checking in proxy side.
	tk := jwt.NewWithClaims(jwt.GetSigningMethod(t.signMethod),
		jwt.MapClaims{
			"username": username,
			"revision": revision,
			"exp":      time.Now().Add(t.ttl).Unix(),
		})

	token, err := tk.SignedString(t.signKey)
	if err != nil {
		if t.lg != nil {
			t.lg.Warn(
				"failed to sign a JWT token",
				zap.String("user-name", username),
				zap.Uint64("revision", revision),
				zap.Error(err),
			)
		} else {
			plog.Debugf("failed to sign jwt token: %s", err)
		}
		return "", err
	}

	if t.lg != nil {
		t.lg.Info(
			"created/assigned a new JWT token",
			zap.String("user-name", username),
			zap.Uint64("revision", revision),
			zap.String("token", token),
		)
	} else {
		plog.Debugf("jwt token: %s", token)
	}
	return token, err
}

func prepareOpts(lg *zap.Logger, opts map[string]string) (jwtSignMethod, jwtPubKeyPath, jwtPrivKeyPath string, ttl time.Duration, err error) {
	for k, v := range opts {
		switch k {
		case "sign-method":
			jwtSignMethod = v
		case "pub-key":
			jwtPubKeyPath = v
		case "priv-key":
			jwtPrivKeyPath = v
		case "ttl":
			ttl, err = time.ParseDuration(v)
			if err != nil {
				if lg != nil {
					lg.Warn(
						"failed to parse JWT TTL option",
						zap.String("ttl-value", v),
						zap.Error(err),
					)
				} else {
					plog.Errorf("failed to parse ttl option (%s)", err)
				}
				return "", "", "", 0, ErrInvalidAuthOpts
			}
		default:
			if lg != nil {
				lg.Warn("unknown JWT token option", zap.String("option", k))
			} else {
				plog.Errorf("unknown token specific option: %s", k)
			}
			return "", "", "", 0, ErrInvalidAuthOpts
		}
	}
	if len(jwtSignMethod) == 0 {
		return "", "", "", 0, ErrInvalidAuthOpts
	}
	return jwtSignMethod, jwtPubKeyPath, jwtPrivKeyPath, ttl, nil
}

func newTokenProviderJWT(lg *zap.Logger, opts map[string]string) (*tokenJWT, error) {
	jwtSignMethod, jwtPubKeyPath, jwtPrivKeyPath, ttl, err := prepareOpts(lg, opts)
	if err != nil {
		return nil, ErrInvalidAuthOpts
	}

	if ttl == 0 {
		ttl = 5 * time.Minute
	}

	t := &tokenJWT{
		lg:  lg,
		ttl: ttl,
	}

	t.signMethod = jwtSignMethod

	verifyBytes, err := ioutil.ReadFile(jwtPubKeyPath)
	if err != nil {
		if lg != nil {
			lg.Warn(
				"failed to read JWT public key",
				zap.String("public-key-path", jwtPubKeyPath),
				zap.Error(err),
			)
		} else {
			plog.Errorf("failed to read public key (%s) for jwt: %s", jwtPubKeyPath, err)
		}
		return nil, err
	}
	t.verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		if lg != nil {
			lg.Warn(
				"failed to parse JWT public key",
				zap.String("public-key-path", jwtPubKeyPath),
				zap.Error(err),
			)
		} else {
			plog.Errorf("failed to parse public key (%s): %s", jwtPubKeyPath, err)
		}
		return nil, err
	}

	signBytes, err := ioutil.ReadFile(jwtPrivKeyPath)
	if err != nil {
		if lg != nil {
			lg.Warn(
				"failed to read JWT private key",
				zap.String("private-key-path", jwtPrivKeyPath),
				zap.Error(err),
			)
		} else {
			plog.Errorf("failed to read private key (%s) for jwt: %s", jwtPrivKeyPath, err)
		}
		return nil, err
	}
	t.signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		if lg != nil {
			lg.Warn(
				"failed to parse JWT private key",
				zap.String("private-key-path", jwtPrivKeyPath),
				zap.Error(err),
			)
		} else {
			plog.Errorf("failed to parse private key (%s): %s", jwtPrivKeyPath, err)
		}
		return nil, err
	}

	return t, nil
}
