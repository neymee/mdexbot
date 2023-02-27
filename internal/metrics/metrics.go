package metrics

import (
	"context"
	"errors"
	"net/http"

	werrors "github.com/neymee/mdexbot/internal/errors"
	"github.com/neymee/mdexbot/internal/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	MessageCounter = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "mdexbot",
		Name:      "message_count_total",
		Help:      "The total number of sent messages",
	})

	errorsCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "mdexbot",
		Name:      "errors_count_total",
		Help:      "The total number of errors",
	}, []string{"error"})

	commandDuration = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "mdexbot",
		Name:      "command_duration_seconds",
		Help:      "The duration of processing messages from users",
	}, []string{"command"})

	httpDuration = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "mdexbot",
		Name:      "http_req_duration_seconds",
		Help:      "The duration of http requests",
	}, []string{"api"})
)

func HandleHTTP(ctx context.Context) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server := http.Server{Addr: ":2112", Handler: mux}
	defer server.Shutdown(ctx)

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Log(ctx, "metrics.HandleHTTP").Err(err).Send()
		}
	}()

	<-ctx.Done()
}

func CommandDuration(cmd string) prometheus.Observer {
	return commandDuration.With(prometheus.Labels{
		"command": cmd,
	})
}

func HTTPDuration(api string) prometheus.Observer {
	return httpDuration.With(prometheus.Labels{
		"api": api,
	})
}

func ErrorsCounter(err error) prometheus.Counter {
	var errLabel string

	switch {
	case errors.Is(err, werrors.DatabaseError):
		errLabel = "database"
	case errors.Is(err, werrors.FailedHTTPReqError):
		errLabel = "http"
	case errors.Is(err, werrors.TelegramError):
		errLabel = "telegram"
	default:
		errLabel = "unknown"
	}

	return errorsCounter.With(prometheus.Labels{
		"error": errLabel,
	})
}
