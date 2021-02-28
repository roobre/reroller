package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
	"path"
	"roob.re/reroller"
	"time"
)

func main() {
	pflag.String("kubeconfig", path.Join(os.ExpandEnv("$HOME"), ".kube", "config"), "path to kubeconfig")
	pflag.String("namespace", "", "namespaces to query (comma-separated)")
	pflag.Bool("unannotated", false, "process unannotated rollouts")
	pflag.Bool("dry-run", false, "do not actually reroll anything")
	pflag.String("log-level", "info", "log level (verbosity)")
	pflag.Duration("interval", 0, "run every [interval], empty to run one. time.ParseDuration format")
	pflag.Duration("cooldown", time.Hour, "do not re-deploy more often than this. time.ParseDuration format")
	pflag.Parse()

	viper.SetEnvPrefix("REROLLER")
	viper.AutomaticEnv()
	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		log.Fatal(err)
	}

	lvl, _ := log.ParseLevel(viper.GetString("log-level"))
	log.SetLevel(lvl)

	rr, err := reroller.NewWithKubeconfig(viper.GetString("kubeconfig"))
	if err != nil {
		log.Fatal(err)
	}
	rr.Unannotated = viper.GetBool("unannotated")
	rr.Namespace = viper.GetString("namespace")
	rr.DryRun = viper.GetBool("dry-run")
	rr.Cooldown = viper.GetDuration("cooldown")

	interval := viper.GetDuration("interval")
	if interval == 0 {
		rr.Run()
		return
	}

	for {
		rr.Run()
		time.Sleep(interval)
	}
}
