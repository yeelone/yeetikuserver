package middleware

import (
	"net/http"
	"regexp"

	c "../config"
	"../utils"
	"github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/urfave/negroni"
	"golang.org/x/net/context"
)

func HeaderMiddleware() negroni.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		// rw.Header().Set("Server", "A Go Web Server")
		rw.Header().Set("Content-Type", "application/json")
		next(rw, r)
	}
}

func CheckAuthMiddleware() negroni.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
			ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
				return []byte(c.Config.SecretKey), nil
			},
			SigningMethod: jwt.SigningMethodHS256,
		})
		token, _ := jwtMiddleware.Options.Extractor(r)
		var ctx context.Context
		if len(token) > 0 {
			id := utils.ParseUserProperty(token)
			ctx = utils.SaveUserInfoToContext(r.Context(), id)
		}
		if m, _ := regexp.MatchString("/home", r.URL.Path); m {
			next(rw, r)
		} else if m, _ := regexp.MatchString("/download", r.URL.Path); m {
			next(rw, r)
		} else if m, _ := regexp.MatchString("/assets", r.URL.Path); m {
			next(rw, r)
		} else if m, _ := regexp.MatchString("/static/", r.URL.Path); m {
			next(rw, r)
		} else if m, _ := regexp.MatchString("/api/v1/auth/register", r.URL.Path); m {
			next(rw, r)
		} else if m, _ := regexp.MatchString("/api/v1/auth/admin/login", r.URL.Path); m {
			next(rw, r)
		} else if m, _ := regexp.MatchString("/api/v1/auth/login", r.URL.Path); !m {
			if jwtMiddleware.CheckJWT(rw, r) != nil {
				rw.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
				http.Error(rw, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			} else {
				next(rw, r.WithContext(ctx))
			}
		} else {
			next(rw, r)
		}
	}
}
