package middleware

import (
	"errors"
	"grampanchayat/models"
	"grampanchayat/utilities"
	"net/http"
)

func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contextValues, ok := r.Context().Value(utilities.UserContextKey).(models.ContextValues)
		if !ok {
			utilities.HandlerError(w, http.StatusInternalServerError, "AdminMiddleware: Context for ID:%v", errors.New("cannot get id from context"))
			return
		}

		if contextValues.Role != "Admin" {
			_, err := w.Write([]byte("ERROR: Role mismatch"))
			if err != nil {
				return
			}

			utilities.HandlerError(w, http.StatusUnauthorized, "Role invalid", errors.New("user role is invalid"))
			return
		}

		next.ServeHTTP(w, r)
	})
}
