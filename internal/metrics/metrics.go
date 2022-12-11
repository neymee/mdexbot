package metrics

import (
	"context"
	"net/http"

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
