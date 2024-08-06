package app

import (
	"flag"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var RootCommand = cobra.Command{
	Use:   "tke-room-manager",
	Short: "A room manager for TKE",
	Run: func(cmd *cobra.Command, args []string) {
		runManager()
	},
}

const (
	metricsBindAddress     = "metrics-bind-address"
	leaderElect            = "leader-elect"
	healthProbeBindAddress = "health-probe-bind-address"
	apiBindAddress         = "api-bind-address"
)

var (
	envReplacer = strings.NewReplacer("-", "_")
	zapOptions  = &zap.Options{
		Development: true,
	}
)

func init() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(envReplacer)

	flags := RootCommand.Flags()
	zapOptions.BindFlags(flag.CommandLine)
	flags.AddGoFlagSet(flag.CommandLine)
	addStringFlag(flags, metricsBindAddress, "0", "The address the metrics endpoint binds to. Use :8443 for HTTPS or :8080 for HTTP, or leave as 0 to disable the metrics service.")
	addStringFlag(flags, healthProbeBindAddress, ":8081", "The address the probe endpoint binds to.")
	addStringFlag(flags, apiBindAddress, ":80", "The api address binds to.")
	addBoolFlag(flags, leaderElect, false, "Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
}

func addStringFlag(flags *pflag.FlagSet, name, value, usage string) {
	flags.String(name, value, wrapUsage(name, usage))
	viper.BindPFlag(name, flags.Lookup(name))
}

func addBoolFlag(flags *pflag.FlagSet, name string, value bool, usage string) {
	flags.Bool(name, value, wrapUsage(name, usage))
	viper.BindPFlag(name, flags.Lookup(name))
}

func wrapUsage(name, usage string) string {
	envName := envReplacer.Replace(name)
	return fmt.Sprintf("%s (ENV: %s)", usage, envName)
}
