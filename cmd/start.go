package cmd

import (
	"fmt"

	"github.com/bestruirui/octopus/internal/conf"
	"github.com/bestruirui/octopus/internal/db"
	"github.com/bestruirui/octopus/internal/op"
	"github.com/bestruirui/octopus/internal/server"
	"github.com/bestruirui/octopus/internal/task"
	"github.com/bestruirui/octopus/internal/utils/log"
	"github.com/bestruirui/octopus/internal/utils/shutdown"
	"github.com/spf13/cobra"
)

var cfgFile string

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start " + conf.APP_NAME,
	PreRun: func(cmd *cobra.Command, args []string) {
		conf.PrintBanner()
		conf.Load(cfgFile)
		log.SetLevel(conf.AppConfig.Log.Level)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		shutdown.Init(log.Logger)
		if err := db.InitDB(conf.AppConfig.Database.Type, conf.AppConfig.Database.Path, conf.IsDebug()); err != nil {
			return fmt.Errorf("initialize database: %w", err)
		}
		shutdown.Register(db.Close)

		if err := op.InitCache(); err != nil {
			shutdown.Shutdown()
			return fmt.Errorf("initialize cache: %w", err)
		}
		shutdown.Register(op.SaveCache)

		if err := op.UserInit(); err != nil {
			shutdown.Shutdown()
			return fmt.Errorf("initialize user: %w", err)
		}

		if err := server.Start(); err != nil {
			shutdown.Shutdown()
			return fmt.Errorf("start server: %w", err)
		}
		shutdown.Register(server.Close)

		task.Init()
		go task.RUN()
		shutdown.Listen()
		return nil
	},
}

func init() {
	startCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./data/config.json)")
	rootCmd.AddCommand(startCmd)
}
