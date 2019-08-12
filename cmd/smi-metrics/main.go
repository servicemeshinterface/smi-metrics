package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/deislabs/smi-metrics/pkg/istio"

	"github.com/deislabs/smi-metrics/pkg/linkerd"
	"github.com/deislabs/smi-metrics/pkg/mesh"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"k8s.io/klog"

	"github.com/deislabs/smi-metrics/pkg/server"
)

var (
	globalUsage = "Expose metrics.smi-spec.io"

	cmd = &cobra.Command{
		Use:   "smi-metrics",
		Short: "Expose metrics.smi-spec.io",
		Long:  globalUsage,
		Run:   run,
	}

	envRoot = "smi-metrics"
)

// bindFlag moves cobra flags into viper for exclusive use there.
func bindFlag(f *pflag.Flag) error {
	v := viper.GetViper()

	if err := v.BindPFlag(f.Name, f); err != nil {
		return err
	}

	if err := v.BindEnv(
		f.Name,
		strings.Replace(
			strings.ToUpper(
				fmt.Sprintf(
					"%s_%s", envRoot, f.Name)), "-", "_", -1)); err != nil {
		return err
	}

	return nil
}

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatalf("unable to execute: %s", err)
	}
}

func cmdFlags(flags *pflag.FlagSet) error {
	flags.String(
		"config",
		"",
		"config file",
	)
	if err := bindFlag(flags.Lookup("config")); err != nil {
		return err
	}

	flags.String(
		"log-level",
		"info",
		"log level to use",
	)
	if err := bindFlag(flags.Lookup("log-level")); err != nil {
		return err
	}

	flags.Int(
		"admin-port",
		8081,
		"port listen on for admin related requests",
	)
	if err := bindFlag(flags.Lookup("admin-port")); err != nil {
		return err
	}

	flags.Int(
		"api-port",
		8080,
		"port listen on for api related requests",
	)
	if err := bindFlag(flags.Lookup("api-port")); err != nil {
		return err
	}

	flags.String(
		"tls-cert-file",
		"/var/run/smi-metrics/tls.crt",
		"TLS certificate to use",
	)
	if err := bindFlag(flags.Lookup("tls-cert-file")); err != nil {
		return err
	}

	flags.String(
		"tls-private-key",
		"/var/run/smi-metrics/tls.key",
		"TLS private key to use",
	)
	if err := bindFlag(flags.Lookup("tls-private-key")); err != nil {
		return err
	}

	flags.String(
		"prometheus-url",
		"http://prometheus.default.svc.cluster.local:9090",
		"URL to use for connecting to prometheus of the format: hostname:port",
	)
	if err := bindFlag(flags.Lookup("prometheus-url")); err != nil {
		return err
	}

	return nil
}

func logConfig() error {
	out, err := yaml.Marshal(viper.AllSettings())
	if err != nil {
		return err
	}

	log.Debugf("Configuration:\n---\n%s---", out)

	return nil
}

func initConfig() {
	flags := cmd.PersistentFlags()
	cfgPath, err := flags.GetString("config")
	if err != nil {
		log.Fatalf("unable to fetch config flag value: %s", err)
	}

	if cfgPath == "" {
		return
	}

	viper.SetConfigFile(cfgPath)

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("unable to read config: %s", err)
	}

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Infof("Config file %s changed", cfgPath)

		if err := logConfig(); err != nil {
			log.Fatalf("unable to marshal and output configuration: %s", err)
		}
	})
}

func initLog() error {
	klog.InitFlags(nil)

	level, err := log.ParseLevel(viper.GetString("log-level"))
	if err != nil {
		return err
	}

	log.SetLevel(level)

	if level == log.DebugLevel {
		if err := flag.Set("stderrthreshold", "INFO"); err != nil {
			return err
		}
		if err := flag.Set("logtostderr", "true"); err != nil {
			return err
		}
		// At 7 and higher, authorization tokens get logged.
		if err := flag.Set("v", "6"); err != nil {
			return err
		}
	}

	return nil
}

//nolint:gochecknoinits
func init() {
	if err := cmdFlags(cmd.PersistentFlags()); err != nil {
		log.Fatalf("unable to parse flags: %s", err)
	}

	cobra.OnInitialize(func() {
		initConfig()
		if err := initLog(); err != nil {
			log.Fatalf("unable to init logging: %s", err)
		}

		if err := logConfig(); err != nil {
			log.Fatalf("unable to marshal and output configuration: %s", err)
		}
	})
}

func run(_ *cobra.Command, args []string) {
	log.Infof("api listening on %d", viper.GetInt("api-port"))
	log.Infof("admin listening on %d", viper.GetInt("admin-port"))

	var meshInstance mesh.Mesh
	var err error
	log.Info("Mesh Config", viper.GetString("mesh"))

	switch provider := viper.GetString("mesh"); provider {
	case "linkerd":
		var config linkerd.Config
		err = viper.UnmarshalKey("linkerd", &config)
		if err != nil {
			log.Fatalf("Unable to unmarshal config into struct")
		}

		meshInstance, err = linkerd.NewLinkerdProvider(config)
		if err != nil {
			log.Fatal("Couldn't create a Linkerd instance", err)
		}
	case "istio":
		var config istio.Config
		err = viper.UnmarshalKey("istio", &config)
		if err != nil {
			log.Fatalf("Unable to unmarshal config into struct")
		}
		meshInstance, err = istio.NewIstioProvider(config)
		if err != nil {
			log.Fatal("Couldn't create a Istio instance", err)
		}
	default:
		log.Fatalf("Unable to recognize the mesh type")

	}

	s := server.Server{
		APIPort:        viper.GetInt("api-port"),
		AdminPort:      viper.GetInt("admin-port"),
		TLSCertificate: viper.GetString("tls-cert-file"),
		TLSPrivateKey:  viper.GetString("tls-private-key"),
		Mesh:           meshInstance,
	}

	if err := s.Listen(); err != nil {
		log.Fatalf("Unable to start listening: %s", err)
	}
}
