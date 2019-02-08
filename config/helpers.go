package config

import (
	"os"
	"os/user"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
)

// InList check if string is in a collection of strings.
func InList(ll []string, n string) bool {
	for _, l := range ll {
		if l == n {
			return true
		}
	}
	return false
}

// InNSList check if ns is in an ns collection.
func InNSList(nn []interface{}, ns string) bool {
	ss := make([]string, len(nn))
	for i, n := range nn {
		ss[i] = n.(v1.Namespace).Name
	}
	return InList(ss, ns)
}

func mustK9sHome() string {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	return usr.HomeDir
}

func mustK9sUser() string {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	return usr.Username
}

func ensurePath(path string, mod os.FileMode) {
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.Mkdir(dir, mod); err != nil {
			log.Errorf("Unable to create K9s home config dir: %v", err)
			panic(err)
		}
	}
}
