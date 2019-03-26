package k8s

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// rbacv1 "k8s.io/client-go/kubernetes/typed/rbac/v1"
)

func GetFu(c Connection, kind, name string) error {
	defer func(t time.Time) {
		log.Info().Msgf("Elapsed %v", time.Since(t))
	}(time.Now())

	crbs, err := c.DialOrDie().RbacV1().ClusterRoleBindings().List(metav1.ListOptions{
		FieldSelector: "metadata.name=cluster-admin",
	})
	if err != nil {
		return err
	}

	log.Info().Msgf("Len %d", len(crbs.Items))

	var crs []string
	for _, crb := range crbs.Items {
		log.Info().Msgf("> CRB %s", crb.Name)
		for _, s := range crb.Subjects {
			// log.Info().Msgf("  Sub %s %s", s.Kind, s.Name)
			if s.Kind == kind && s.Name == name {
				crs = append(crs, crb.RoleRef.Name)
			}
		}
	}

	fmt.Printf("Find cluster roles %#v\n", crs)

	// Each role has multiple rules
	for _, r := range crs {
		cr, err := c.DialOrDie().RbacV1().ClusterRoles().Get(r, metav1.GetOptions{})
		if err != nil {
			log.Error().Err(err).Msgf("Unable to find cluster role %s ", r)
		}
		for _, rule := range cr.Rules {
			log.Info().Msgf("Found rule %#v", rule.APIGroups)
		}
	}

	return nil
}
