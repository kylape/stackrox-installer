package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kylape/stackrox-installer/installer/manifest"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func log(msg string, params ...interface{}) {
	fmt.Printf(msg+"\n", params...)
}

func printHelp() {
	fmt.Println("StackRox Installer")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  installer [options] <action> <set>")
	fmt.Println()
	fmt.Println("Actions:")
	fmt.Println("  apply   - Apply manifests to Kubernetes cluster")
	fmt.Println("  export  - Export manifests to stdout")
	fmt.Println()
	fmt.Println("Sets:")
	fmt.Println("  central         - Central components")
	fmt.Println("  securedcluster  - Secured cluster components")
	fmt.Println("  crs             - Custom resource definitions")
	fmt.Println()
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  installer apply central")
	fmt.Println("  installer export securedcluster")
	fmt.Println("  installer -conf config.yaml apply crs")
}

func main() {
	configPath := flag.String("conf", "./installer.yaml", "Path to installer's configuration file.")
	// kubeconfig = flag.String("kubeconfig", os.Getenv("KUBECONFIG"), "(optional) absolute path to the kubeconfig file")
	// kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	kubeconfigFlag := flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	flag.Parse()

	action := flag.Arg(0)
	generatorSet := flag.Arg(1)

	if action == "" || generatorSet == "" {
		printHelp()
		return
	}

	cfg, err := manifest.ReadConfig(*configPath)
	if err != nil {
		fmt.Printf("failed to load configuration %q: %v\n", *configPath, err)
		return
	}

	cfg.Action = action

	var config *rest.Config
	var clientset *kubernetes.Clientset

	// Only initialize Kubernetes client for apply operations
	if action == "apply" {
		kubeconfig := os.Getenv("KUBECONFIG")

		if kubeconfig == "" {
			kubeconfig = *kubeconfigFlag
		}

		if kubeconfig == "" {
			config, err = rest.InClusterConfig()
			if err != nil {
				home := homedir.HomeDir()
				config, err = clientcmd.BuildConfigFromFlags("", filepath.Join(home, ".kube", "config"))
				if err != nil {
					println(err.Error())
					return
				}
			}
		} else {
			config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
			if err != nil {
				println(err.Error())
				return
			}
		}

		clientset, err = kubernetes.NewForConfig(config)
		if err != nil {
			println(err.Error())
			return
		}
	}

	ctx := context.Background()

	m, err := manifest.New(cfg, clientset, config)
	if err != nil {
		println(err.Error())
		return
	}

	set, found := manifest.GeneratorSets[generatorSet]
	if !found {
		fmt.Printf("Invalid set '%s'. Valid options are central, securedcluster, or crs\n", generatorSet)
		return
	}

	switch action {
	case "apply":
		if err = m.Apply(ctx, *set); err != nil {
			println(err.Error())
			return
		}
	case "export":
		if err = m.Export(ctx, *set); err != nil {
			println(err.Error())
			return
		}
	}
}
