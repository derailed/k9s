package k8s

import (
	"strings"

	"github.com/rs/zerolog"
	authorizationv1 "k8s.io/api/authorization/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
)

// CanIAccess checks if user has access to a certain resource.
func CanIAccess(cfg *Config, log zerolog.Logger, ns, verb, name, resURL string) bool {
	_, gr := schema.ParseResourceArg(strings.ToLower(resURL))
	sar := &authorizationv1.SelfSubjectAccessReview{
		Spec: authorizationv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authorizationv1.ResourceAttributes{
				Namespace:   ns,
				Verb:        verb,
				Group:       gr.Group,
				Resource:    gr.Resource,
				Subresource: "",
				Name:        name,
			},
		},
	}

	rest, err := cfg.RESTConfig()
	if err != nil {
		log.Warn().Msgf("Access %s", err)
		return false
	}

	auth, err := kubernetes.NewForConfig(rest)
	if err != nil {
		log.Warn().Msgf("Access %s", err)
		return false
	}

	response, err := auth.AuthorizationV1().SelfSubjectAccessReviews().Create(sar)
	if err != nil {
		log.Warn().Msgf("Access %s", err)
		return false
	}

	return response.Status.Allowed
}
