package k8s

import (
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func TestRBACFu(t *testing.T) {
	con := dial()
	assert.Nil(t, GetFu(con, "Group", "system:masters"))
}

func dial() *APIClient {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// c, u := "gke_k9s", "gke_k9s_user"
	c, u := "minikube", "minikube"

	flags := genericclioptions.ConfigFlags{
		ClusterName:  &c,
		AuthInfoName: &u,
	}
	cfg := NewConfig(&flags)
	return InitConnectionOrDie(cfg, log.Logger)
}
