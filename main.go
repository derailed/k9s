// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package main

import (
	"flag"
	"os"

	"github.com/derailed/k9s/cmd"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/klog/v2"
)

func init() {
	klog.InitFlags(nil)

	var logFile string
	for i, a := range os.Args {
		if a == "--logFile" && i+1 < len(os.Args) {
			logFile = os.Args[i+1]
			break
		}
	}
	if logFile != "" {
		if err := flag.Set("log_file", logFile); err != nil {
			panic(err)
		}
	}
	if err := flag.Set("logtostderr", "false"); err != nil {
		panic(err)
	}
	if err := flag.Set("alsologtostderr", "false"); err != nil {
		panic(err)
	}
	if err := flag.Set("stderrthreshold", "fatal"); err != nil {
		panic(err)
	}
	if err := flag.Set("v", "-10"); err != nil {
		panic(err)
	}
}

func main() {
	cmd.Execute()
}
