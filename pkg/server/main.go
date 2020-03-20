package server

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/hellofresh/health-go"
	http_logrus "github.com/improbable-eng/go-httpwares/logging/logrus"
	http_metrics "github.com/improbable-eng/go-httpwares/metrics"
	http_prometheus "github.com/improbable-eng/go-httpwares/metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/servicemeshinterface/smi-metrics/pkg/cluster"
	"github.com/servicemeshinterface/smi-metrics/pkg/mesh"
	"github.com/servicemeshinterface/smi-metrics/pkg/metrics"
	metricsAPI "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/metrics/v1alpha2"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Server provides the HTTP serving functionality
type Server struct {
	APIPort   int
	AdminPort int

	TLSCertificate string
	TLSPrivateKey  string

	Mesh mesh.Mesh

	clientNames map[string]bool
	// Used for error messaging in authorizer
	clientNamesOriginal string
}

func (s *Server) getDefaultRouter() *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.StripSlashes)
	router.Use(middleware.GetHead)
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)
	router.Use(http_logrus.Middleware(log.NewEntry(log.StandardLogger())))

	router.Use(
		http_metrics.Middleware(
			http_prometheus.ServerMetrics(
				http_prometheus.WithLatency())))

	return router
}

func (s *Server) adminRouter() *chi.Mux {
	router := s.getDefaultRouter()

	router.Mount("/debug", middleware.Profiler())
	router.Mount("/metrics", promhttp.Handler())
	router.Mount("/status", health.Handler())

	return router
}

// APIRouter encapsulates all the routes required to serve the API
func (s *Server) APIRouter() (*chi.Mux, error) {
	router := s.getDefaultRouter()

	handler, err := metrics.NewHandler(s.Mesh)
	if err != nil {
		return nil, err
	}

	router.Route(
		fmt.Sprintf("/apis/%s", metricsAPI.APIVersion),
		func(router chi.Router) {
			router.Use(s.authorizer)

			router.Mount("/", handler.Routes())
		})

	return router, nil
}

func (s *Server) getTLSConfig() (*tls.Config, error) {
	client, err := cluster.GetClient()
	if err != nil {
		return nil, err
	}

	authConfig, err := client.CoreV1().ConfigMaps("kube-system").Get(
		"extension-apiserver-authentication", metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf(
			"unable to fetch extension-apiserver-authentication: %s", err)
	}
	log.Debug("fetched extension-apiserver-authentication")

	clientCA := authConfig.Data["requestheader-client-ca-file"]

	clientBundle := x509.NewCertPool()
	if ok := clientBundle.AppendCertsFromPEM([]byte(clientCA)); !ok {
		return nil, fmt.Errorf("unable to create client CA bundle")
	}

	var names []string
	s.clientNamesOriginal = authConfig.Data["requestheader-allowed-names"]

	if err := json.Unmarshal([]byte(s.clientNamesOriginal), &names); err != nil {
		return nil, err
	}

	s.clientNames = make(map[string]bool)
	for _, v := range names {
		s.clientNames[v] = true
	}

	return &tls.Config{
		ClientAuth: tls.VerifyClientCertIfGiven,
		ClientCAs:  clientBundle,
	}, nil
}

// Listen starts listening for incoming requests for the APIService and admin
// handlers
func (s *Server) Listen() error {
	go func() {
		if err := http.ListenAndServe(
			fmt.Sprintf("0.0.0.0:%d", s.AdminPort),
			s.adminRouter()); err != nil {
			log.Fatalf("Failed to serve admin routes: %s", err)
		}
	}()

	tlsConfig, err := s.getTLSConfig()
	if err != nil {
		return err
	}

	apiRouter, err := s.APIRouter()
	if err != nil {
		return err
	}

	httpServer := &http.Server{
		Addr:      fmt.Sprintf("0.0.0.0:%d", s.APIPort),
		TLSConfig: tlsConfig,
		Handler:   apiRouter,
	}

	return httpServer.ListenAndServeTLS(
		s.TLSCertificate,
		s.TLSPrivateKey)
}
