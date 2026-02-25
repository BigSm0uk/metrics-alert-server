package middleware

import (
	"net"
	"net/http"
)

func WithSubnetCheck(trustedSubnet string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if trustedSubnet == "" {
				next.ServeHTTP(w, r)
				return
			}

			_, ipNet, err := net.ParseCIDR(trustedSubnet)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			clientIP := r.Header.Get("X-Real-IP")
			if clientIP == "" {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			ip := net.ParseIP(clientIP)
			if ip == nil {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			if !ipNet.Contains(ip) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
