package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

type Authorization struct {
	token string
}

func NewAuthorization(token string) Authorization {
	return Authorization{
		token: token,
	}
}

func (m *Authorization) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(m.token) > 0 {
			token, err := m.getToken(r.Header)
			if err != nil {
				logrus.Error(fmt.Sprintf("error while reading token: %v", err))
				respondWithError(w, ErrHTTPForbidden)
				return
			}

			if m.token != token {
				logrus.Error("invalid token received")
				respondWithError(w, ErrHTTPForbidden)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (m *Authorization) getToken(header http.Header) (string, error) {
	authHeader := "Authorization"
	bearerHeader := "bearer "

	reqToken := header.Get(authHeader)
	if len(reqToken) < len(bearerHeader) {
		return "", errors.New("invalid token provided")
	}

	bearer := strings.ToLower(reqToken[:len(bearerHeader)])
	if bearer != bearerHeader {
		return "", errors.New("invalid token provided")
	}

	token := strings.TrimSpace(reqToken[len(bearerHeader):])

	return token, nil
}

func respondWithError(w http.ResponseWriter, err string) {
	respond(w, http.StatusInternalServerError, err)
}

func respond(w http.ResponseWriter, httpStatus int, responseBody ...interface{}) {
	w.WriteHeader(httpStatus)

	if len(responseBody) > 0 {
		resBytes, err := json.Marshal(responseBody[0])
		if err != nil {
			respondWithError(w, ErrHTTPInternalServerError)
			return
		}
		if _, err = w.Write(resBytes); err != nil {
			logrus.Error(fmt.Sprintf("error while reading token: %v", err))
			respondWithError(w, ErrHTTPInternalServerError)
			return
		}
	}
}
