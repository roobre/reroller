package main

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
	"roob.re/reroller"
	"strconv"
)

func main() {
	kubeconfig := flag.String(
		"kubeconfig",
		path.Join(os.ExpandEnv("$HOME"), ".kube", "config"),
		"path to .kube/config, if running outside the cluster",
	)
	namespace := flag.String("namespace", os.ExpandEnv("REROLLER_NAMESPACE"), "Namespace, defaults to all")
	unannotatedDefault, _ := strconv.ParseBool(os.ExpandEnv("REROLLER_UNANNOTATED"))
	unannotated := flag.Bool("unannotated", unannotatedDefault, "process unannotated pods as well")
	dryRun := flag.Bool("dry-run", false, "do not redeploy anything")
	debuglvl := flag.String("debuglvl", os.ExpandEnv("REROLLER_DEBUG_LVL"), "debug level")
	flag.Parse()

	lvl, _ := log.ParseLevel(*debuglvl)
	log.SetLevel(lvl)

	rr, err := reroller.NewWithKubeconfig(*kubeconfig)
	if err != nil {
		log.Fatal(err)
	}
	rr.Unannotated = *unannotated
	rr.Namespace = *namespace
	rr.DryRun = *dryRun

	rr.Run()
}
