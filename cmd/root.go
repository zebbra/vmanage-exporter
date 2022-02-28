package cmd

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/zebbra/vmanage-exporter/internal/lib/collector"
	"github.com/zebbra/vmanage-exporter/internal/lib/version"
	"github.com/zebbra/vmanage-exporter/internal/lib/vmanage"
	"go.uber.org/zap"
	"net/http"
	"os"
	"time"
)

const userEnv = "VMANAGE_USER"
const passwordEnv = "VMANAGE_PASSWORD"

var rootCmd = &cobra.Command{
	Use:           "vmanage-exporter --vmanage.endpoint <url>",
	SilenceErrors: true,
	Version:       fmt.Sprintf("%s-%s", version.Version, version.Commit),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if os.Getenv(userEnv) == "" || os.Getenv(passwordEnv) == "" {
			return fmt.Errorf(
				"Please provide vmanage credentials in environment variables %s and %s",
				userEnv,
				passwordEnv,
			)
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		endpoint, err := cmd.Flags().GetString("vmanage.endpoint")

		if err != nil {
			return err
		}

		addr, err := cmd.Flags().GetString("web.listen-address")

		if err != nil {
			return err
		}

		metricsPath, err := cmd.Flags().GetString("web.metrics-path")

		if err != nil {
			return err
		}

		scrapeInterval, err := cmd.Flags().GetDuration("scrape.interval")

		if err != nil {
			return err
		}

		logger, _ := zap.NewProduction()
		defer logger.Sync()
		sugar := logger.Sugar()

		sugar.Infof("Validate login on %s", endpoint)

		vmClient := vmanage.NewClient(
			endpoint,
			os.Getenv(userEnv),
			os.Getenv(passwordEnv),
		)

		if v, _ := cmd.Flags().GetBool("tls.verify"); !v {
			vmClient.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		}

		err = vmClient.Login()

		if err != nil {
			return fmt.Errorf("Initial login to %s failed: %w", endpoint, err)
		}

		//o := &vmanage.DeviceInterfaceListOptions{DeviceID: "10.10.1.5"}

		//devices, err := vmClient.Device(context.Background())
		//
		//if err != nil {
		//	return err
		//}
		//
		//litter.Dump(devices)
		//_ = devices
		//
		//err = vmClient.Logout()
		//if err != nil {
		//	return err
		//}
		//
		//return nil

		reg := prometheus.NewPedanticRegistry()

		mainCache := cache.New(5*scrapeInterval, 10*scrapeInterval)

		sugar.Infof("Start initial data collection")

		ctx := context.Background()
		errorCounter := collector.Counter(0)

		sc := &collector.StatisticsCollector{
			Logger:       sugar,
			Cache:        mainCache,
			ErrorCounter: &errorCounter,
		}

		_ = sc.Run(ctx)
		_ = reg.Register(sc)

		vc := &collector.VmanageCollector{
			Logger:       sugar,
			Client:       vmClient,
			Cache:        mainCache,
			ErrorCounter: &errorCounter,
		}

		_ = vc.Run(ctx)
		_ = reg.Register(vc)

		sugar.Infof("Start collector threads")
		scraper := time.NewTicker(scrapeInterval + scrapeInterval/2)
		go func() {
			for {
				select {
				case <-scraper.C:
					go func() {
						ctx, cancel := context.WithTimeout(ctx, scrapeInterval)
						defer cancel()
						_ = vc.Run(ctx)
					}()
				}
			}
		}()

		http.Handle(metricsPath, promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			max, _ := cmd.Flags().GetInt("scrape.max-errors")

			if errorCounter.Get() > int64(max) {
				w.WriteHeader(500)
				_, _ = w.Write([]byte("Unhealthy"))
				return
			}

			_, _ = w.Write([]byte("OK"))
		})

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`
            <html>
            <head><title>vManage Exporter Metrics</title></head>
            <body>
            <h1>Metrics</h1>
            <p><a href='` + metricsPath + `'>Metrics</a></p>
            </body>
            </html>
        `))
		})

		sugar.Infof("Start listening for connections on %s", addr)
		return http.ListenAndServe(addr, nil)
	},
}

// Execute runs root command
func Execute() {
	rootCmd.Flags().String("vmanage.endpoint", "", "URL of vManage API")
	_ = rootCmd.MarkFlagRequired("vmanage.endpoint")

	rootCmd.Flags().String("web.listen-address", ":9910", "Address on which to expose metrics and web interface.")
	rootCmd.Flags().String("web.metrics-path", "/metrics", "Path under which to expose metrics.")

	rootCmd.Flags().Bool("tls.verify", true, "Verify certificate.")

	rootCmd.Flags().Duration("scrape.interval", 15*time.Second, "Polling interval")
	rootCmd.Flags().Int("scrape.max-errors", 100, "Max scrape errors before reporting exporter as unhealthy")

	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
