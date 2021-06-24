package server

import (
	"github.com/patrick246/shortlink/pkg/persistence"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"regexp"
	"time"
)

var codePathRegex = regexp.MustCompile("^/([^/]+)$")

var codeUsageCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "shortlink_code_request_count",
	Help: "Counts the number of requests for a shortcode",
}, []string{"shortcode"})

func (s *Server) handleCodeRequests(w http.ResponseWriter, r *http.Request) {
	matches := codePathRegex.FindStringSubmatch(r.URL.Path)
	if len(matches) != 2 {
		http.Error(w, "Not found", 404)
		return
	}
	code := matches[1]

	shortLink, err := s.repo.GetEntryForCode(r.Context(), code)
	if err == persistence.ErrNotFound {
		log.Warnw("invalid code", "code", code, "ip", r.RemoteAddr)
		http.Error(w, "Not found", 404)
		return
	}
	if err != nil {
		log.Errorw("error getting code", "code", code, "error", err, "ip", r.RemoteAddr)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	if !shortLink.TTL.IsZero() && shortLink.TTL.Before(time.Now()) {
		http.Error(w, "Not found", 404)
		return
	}

	codeUsageCounter.WithLabelValues(code).Inc()
	http.Redirect(w, r, shortLink.URL, http.StatusFound)
}
