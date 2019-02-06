package views

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"github.com/derailed/k9s/resource"
	"github.com/derailed/k9s/resource/k8s"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
)

const (
	defaultRefreshRate   = 2
	defaultLogBufferSize = 200
	defaultView          = "po"
)

var (
	k9sCfg = defaultK9sConfig()
	// K9sHome represent K9s home directory.
	K9sHome = filepath.Join(mustK9sHome(), ".k9s")
	// K9sConfig represents K9s config file location.
	K9sConfig = filepath.Join(K9sHome, "config.yml")
	// K9sLogs represents K9s log.
	K9sLogs = filepath.Join(os.TempDir(), fmt.Sprintf("k9s-%s.log", mustK9sUser()))
)

const maxFavorites = 5

type (
	namespace struct {
		Active    string   `yaml:"active"`
		Favorites []string `yaml:"favorites"`
	}

	view struct {
		Active string `yaml:"active"`
	}

	k9s struct {
		RefreshRate   int       `yaml:"refreshRate"`
		LogBufferSize int       `yaml:"logBufferSize"`
		Namespace     namespace `yaml:"namespace"`
		View          view      `yaml:"view"`
	}

	config struct {
		K9s k9s `yaml:"k9s"`
	}
)

func (c *config) load(path string) {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		log.Errorf("[Config] Unable to locate K9s configuration in %s. Using defaults", path)
	}

	if err = yaml.Unmarshal(f, k9sCfg); err != nil {
		log.Errorf("[Config] hydrating K9s YAML: %v", err)
	}
	log.Debugf("[Config] Loaded K9s configuration `%s", c.K9s.Namespace.Active)
}

func (c *config) save(path string) error {
	log.Debugf("[Config] Saving configuration `%s", c.K9s.Namespace.Active)
	ensurePath(path, 0755)
	cfg, err := yaml.Marshal(c)
	if err != nil {
		log.Errorf("[Config] Unable to save K9s config file: %v", err)
		return err
	}
	return ioutil.WriteFile(path, cfg, 0644)
}

func (c *config) validateAndSave() error {
	c.validate()
	return c.save(K9sConfig)
}

func (c *config) validate() {
	if c.K9s.RefreshRate <= 0 {
		c.K9s.RefreshRate = defaultRefreshRate
	}

	if c.K9s.LogBufferSize <= 0 {
		c.K9s.LogBufferSize = defaultLogBufferSize
	}

	nn, err := activeNamespaces()
	if err != nil {
		return
	}
	if !c.isAllNamespace() && !inList(nn, c.K9s.Namespace.Active) {
		log.Debugf("[Config] Validation error active namepace reset to `default")
		c.K9s.Namespace.Active = resource.DefaultNamespace
	}
	for _, f := range c.K9s.Namespace.Favorites {
		if f != resource.AllNamespace && !inList(nn, f) {
			log.Debugf("[Config] Invalid favorite found '%s' - %t", f, c.isAllNamespace())
			c.rmFavNS(f)
		}
	}
}

func (c *config) isAllNamespace() bool {
	return c.K9s.Namespace.Active == resource.AllNamespace
}

func (*config) reset() {
	k9sCfg = defaultK9sConfig()
}

func (c *config) addActive(ns string) {
	c.K9s.Namespace.Active = ns
	c.addFavNS(ns)
}

func (c *config) addFavNS(ns string) {
	fv := c.K9s.Namespace.Favorites
	if inList(fv, ns) {
		return
	}

	nfv := make([]string, 0, maxFavorites)
	nfv = append(nfv, ns)
	for i := 0; i < len(fv); i++ {
		if i+1 < maxFavorites {
			nfv = append(nfv, fv[i])
		}
	}
	c.K9s.Namespace.Favorites = nfv
}

func (c *config) rmFavNS(ns string) {
	fv, victim := c.K9s.Namespace.Favorites, -1
	for i, f := range fv {
		if f == ns {
			victim = i
			break
		}
	}

	if victim < 0 {
		return
	}
	fv = append(fv[:victim], fv[victim+1:]...)
	c.K9s.Namespace.Favorites = fv
}

func defaultK9sConfig() *config {
	return &config{
		K9s: k9s{
			RefreshRate:   5,
			LogBufferSize: 200,
			View: view{
				Active: "po",
			},
			Namespace: namespace{
				Active: resource.DefaultNamespace,
				Favorites: []string{
					resource.AllNamespace,
					resource.DefaultNamespace,
					"kube-system",
				},
			},
		},
	}
}

func inList(ll []string, n string) bool {
	for _, l := range ll {
		if l == n {
			return true
		}
	}
	return false
}

func inNSList(nn []interface{}, ns string) bool {
	for _, n := range nn {
		nsp := n.(v1.Namespace)
		if nsp.Name == ns {
			return true
		}
	}
	return false
}

func activeNamespaces() ([]string, error) {
	nn, err := k8s.NewNamespace().List(defaultNS)
	if err != nil {
		log.Errorf("Unable to retrieve active namespaces: %#v", err.Error())
		return []string{}, err
	}

	ss := make([]string, len(nn))
	for i, n := range nn {
		ss[i] = n.(v1.Namespace).Name
	}
	return ss, nil
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
