package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"grampanchayat/database/helper"
	"grampanchayat/handler"
	"grampanchayat/models"
	"grampanchayat/utilities"
	"net/http"
	"runtime/debug"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("token")

		claims := &models.Claims{}

		tkn, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
			return handler.JwtKey, nil
		})

		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				utilities.HandlerError(w, http.StatusUnauthorized, "Signature invalid:%v", err)
				return
			}
			utilities.HandlerError(w, http.StatusBadRequest, "ParseError:%v", err)
			return
		}

		if tkn == nil {
			utilities.HandlerError(w, http.StatusUnauthorized, "unable to find token", fmt.Errorf("no token found"))
			return
		}

		if !tkn.Valid {
			utilities.HandlerError(w, http.StatusUnauthorized, "token is invalid:", errors.New("token is invalid"))
			return
		}

		_, err = helper.CheckSession(claims.ID)
		if err != nil {
			logrus.Printf("session expired:%v", err)
			return
		}

		value := &models.ContextValues{ID: claims.ID, Role: claims.Role}
		ctx := context.WithValue(r.Context(), utilities.UserContextKey, *value)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// corsOptions setting up routes for cors
func corsOptions() *cors.Cors {
	return cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "Access-Token", "importDate", "X-Client-Version", "Cache-Control", "Pragma", "x-started-at", "x-api-key", "token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
	})
}

// CommonMiddlewares middleware common for all routes
func CommonMiddlewares() chi.Middlewares {
	return chi.Chain(
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("Content-Type", "application/json")
				next.ServeHTTP(w, r)
			})
		},
		corsOptions().Handler,
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer func() {
					err := recover()
					if err != nil {
						logrus.Errorf("Request Panic err: %v", err)
						jsonBody, _ := json.Marshal(map[string]string{
							"error": "There was an internal server error",
							"trace": fmt.Sprintf("%+v", err),
							"stack": string(debug.Stack()),
						})
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusInternalServerError)
						_, err := w.Write(jsonBody)
						if err != nil {
							logrus.Errorf("Failed to send response from middleware with error: %+v", err)
						}
					}
				}()
				next.ServeHTTP(w, r)
			})
		},
	)
}
