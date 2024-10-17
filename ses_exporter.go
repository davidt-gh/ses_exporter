package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promslog"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
)

var (
	metricsPath = kingpin.Flag(
		"web.telemetry-path",
		"Path under which to expose metrics.",
	).Default("/metrics").String()
	toolkitFlags = webflag.AddFlags(kingpin.CommandLine, ":9101")

	// Metrics
	sesMax24HourSend = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ses_quota_max",
		Help: "The maximum number of emails the user is allowed to send in a 24-hour interval.",
	})
	sesMaxSendRate = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ses_quota_rate",
		Help: "The maximum number of emails that Amazon SES can accept from the user's account per second.",
	})
	sesSentLast24Hours = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ses_quota_sent",
		Help: "The number of emails sent during the previous 24 hours.",
	})

	client *ses.SES
	logger *slog.Logger
)

// Periodically retrieve the SES metrics using the AWS client.
func recordMetrics() {
	go func() {
		for {

			// Retrieve the current quota.
			result, err := client.GetSendQuota(nil)
			if err != nil {
				logger.Error("Error retrieving the sending limits for the Amazon SES account", "err", err)
			}

			// Update the gauges.
			sesMax24HourSend.Set(aws.Float64Value(result.Max24HourSend))
			sesMaxSendRate.Set(aws.Float64Value(result.MaxSendRate))
			sesSentLast24Hours.Set(aws.Float64Value(result.SentLast24Hours))

			// Sleep for 5 seconds.
			time.Sleep(5 * time.Second)
		}
	}()
}

func main() {
	promslogConfig := &promslog.Config{}
	kingpin.Version(version.Print("ses_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger = promslog.New(promslogConfig)

	logger.Info("Starting ses_exporter", "version", version.Info())
	logger.Info("operational information", "build_context", version.BuildContext())

	// Initialize a session that the SDK uses to load
	// credentials from the shared credentials file ~/.aws/credentials
	// and configuration from the shared configuration file ~/.aws/config.
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create an SES client with a session.
	client = ses.New(sess)

	recordMetrics()

	// Define the metrics page based on the metrics path.
	http.Handle(*metricsPath, promhttp.Handler())

	// Define the landing page if the metrics path is not set to the root.
	if *metricsPath != "/" {
		landingConfig := web.LandingConfig{
			Name:        "AWS SES Exporter",
			Description: "Amazon Simple Email Service Exporter",
			Version:     version.Info(),
			Links: []web.LandingLinks{
				{
					Address: *metricsPath,
					Text:    "Metrics",
				},
			},
		}
		landingPage, err := web.NewLandingPage(landingConfig)
		if err != nil {
			logger.Error(err.Error())
			os.Exit(1)
		}
		http.Handle("/", landingPage)
	}

	// Start the HTTP server.
	srv := &http.Server{}
	if err := web.ListenAndServe(srv, toolkitFlags, logger); err != nil {
		logger.Error("Error starting HTTP server", "err", err)
		os.Exit(1)
	}
}
