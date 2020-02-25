package server

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

var (
	unauthorized = "Unauthorized client certificate. Check configuration and try again."
)

func (s *Server) authorizer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.TLS.PeerCertificates) == 0 {
			log.Infof("missing client certificate")
			http.Error(w, unauthorized, http.StatusUnauthorized)
			return
		}

		if len(s.clientNames) == 0 {
			next.ServeHTTP(w, r)
			return
		}
		for _, crt := range r.TLS.PeerCertificates {
			name := crt.Subject.CommonName
			if _, ok := s.clientNames[name]; ok {
				next.ServeHTTP(w, r)

				return
			}

			log.Infof(
				"invalid client certificate name: %s, must be one of: %s",
				name,
				s.clientNamesOriginal)
		}

		http.Error(w, unauthorized, http.StatusUnauthorized)
	})
}
