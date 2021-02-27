package main

import (
	"flag"
	"log"
	"os"
	"path"
	"roob.re/reroller"
)

func main() {
	kubeconfig := flag.String(
		"kubeconfig",
		path.Join(os.ExpandEnv("$HOME"), ".kube", "config"),
		"path to .kube/config, if running outside the cluster",
	)
	namespace := flag.String("namespace", "", "Namespace, defaults to all")
	unannotated := flag.Bool("unannotated", false, "process unnanotated pods as well")
	flag.Parse()

	rr, err := reroller.NewWithKubeconfig(*kubeconfig)
	if err != nil {
		log.Fatal(err)
	}
	rr.Unannotated = *unannotated
	rr.Namespace = *namespace

	rr.Run()
}
