package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/joshinjohnson/wave-exercise/handler"
	"github.com/joshinjohnson/wave-exercise/pkg/db"
	"github.com/joshinjohnson/wave-exercise/pkg/payroll"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	appName        = "payroll-api"
	defaultTimeout = 30 * time.Minute
	failureCount   = 3
)

var (
	addr    string
	cfgFile string
	cfg     *Config
)

var rootCmd = &cobra.Command{
	Use:   "payroll",
	Short: "Employee Payroll Generator API",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if cfg == nil {
			return errors.New("error while reading config file")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		log.Info("setting up dependencies")

		dbConfig := make(map[string]string, 0)
		dbConfig[db.UsernameField] = cfg.DbConfig.User
		dbConfig[db.PasswordField] = cfg.DbConfig.Password
		dbConfig[db.HostNameField] = cfg.DbConfig.Hostname
		dbConfig[db.PortField] = cfg.DbConfig.Port
		dbConfig[db.DbNameField] = cfg.DbConfig.DatabaseName
		dbConfig[db.SchemaField] = cfg.DbConfig.SchemaName
		dbConfig[db.SSLModeField] = cfg.DbConfig.SSLMode

		dbW, err := db.NewDbWrapper(ctx, dbConfig)
		if err != nil {
			log.Errorf("error while setting up db client: %v", err)
			os.Exit(1)
		}
		defer dbW.DB.Close()

		payrollService := payroll.NewPayrollService(dbW)
		payrollHandler := handler.NewPayrollHandler(payrollService)
		authMiddleware := handler.NewAuthorization(cfg.AuthToken)
		contextMiddleware := handler.NewContext()

		h := handler.Handler(payrollHandler)
		h = removeTrailingSlash(h)
		h = addDefaultHeader(h)
		h = contextMiddleware.Middleware(h)
		h = authMiddleware.Middleware(h)

		apiServer := http.Server{
			Addr:         cfg.ServerAddress,
			Handler:      h,
			IdleTimeout:  defaultTimeout,
			ReadTimeout:  defaultTimeout,
			WriteTimeout: defaultTimeout,
		}

		log.Info("finished setting up dependencies, starting services")

		var wg sync.WaitGroup
		wg.Add(1)
		startAPIHandler(&wg, &apiServer)
		wg.Wait()

		log.Info("successfully stopped payroll-api")
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config file path", "", "config file (default is $HOME/payroll_config.yaml)")
	log.SetLevel(log.InfoLevel)
}

func initConfig() {
	viper.SetConfigName(appName)

	var err error
	cfg, err = LoadConfig(cfgFile)
	if err != nil {
		log.Fatalf("error while reading config file: %v", err)
		os.Exit(1)
	}
}

func removeTrailingSlash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		request.URL.Path = strings.TrimSuffix(request.URL.Path, "/")
		next.ServeHTTP(writer, request)
	})
}

func addDefaultHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(writer, request)
	})
}

func startAPIHandler(wg *sync.WaitGroup, server *http.Server) {
	log.Info(fmt.Sprintf("started listening on: '%v'", server.Addr))
	defer wg.Done()
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error(fmt.Sprintf("could not listen on %s: %v", cfg.ServerAddress, err))
		return
	}
}
