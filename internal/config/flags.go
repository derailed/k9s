package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	// DefaultRefreshRate represents the refresh interval.
	DefaultRefreshRate = 2 // secs

	// DefaultLogLevel represents the default log level.
	DefaultLogLevel = "info"

	// DefaultCommand represents the default command to run.
	DefaultCommand = ""

	// KubeconfigDirDefault represents the default kubeconfig directory.
	KubeconfigDirDefault = ""
)

// Flags represents K9s configuration flags.
type Flags struct {
	RefreshRate   *int
	LogLevel      *string
	Headless      *bool
	Command       *string
	AllNamespaces *bool
	ReadOnly      *bool
	Write         *bool
	Crumbsless    *bool
	KubeconfigDir *string
}

// NewFlags returns new configuration flags.
func NewFlags() *Flags {
	return &Flags{
		RefreshRate:   intPtr(DefaultRefreshRate),
		LogLevel:      strPtr(DefaultLogLevel),
		Headless:      boolPtr(false),
		Command:       strPtr(DefaultCommand),
		AllNamespaces: boolPtr(false),
		ReadOnly:      boolPtr(false),
		Write:         boolPtr(false),
		Crumbsless:    boolPtr(false),
		KubeconfigDir: strPtr(KubeconfigDirDefault),
	}
}

func boolPtr(b bool) *bool {
	return &b
}

func intPtr(i int) *int {
	return &i
}

func strPtr(s string) *string {
	return &s
}

// IsKubeconfigDirSet returns true if the kubeconfigDir was set and false otherwise.
func (f *Flags) IsKubeconfigDirSet() bool {
	return *f.KubeconfigDir != ""
}

// Kubeconfig returns the kubeconfig of choice.
func (f *Flags) Kubeconfig() string {

	var files []string
	err := filepath.Walk(*f.KubeconfigDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return ""
	}

	textInt := chooseKubeconfig(files)

	return files[textInt]
}

func chooseKubeconfig(files []string) int {
	for index, file := range files {
		fmt.Printf("%d:\t%s\n", index, file)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Choose the config: ")
	text, _ := reader.ReadString('\n')
	textInt, _ := strconv.Atoi(strings.Replace(text, "\n", "", -1))

	return textInt
}
