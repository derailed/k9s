package k8s

import (
	"strings"

	"github.com/rs/zerolog/log"
	authorizationv1 "k8s.io/api/authorization/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
)

// CanIAccess checks if user has access to a certain resource.
func CanIAccess(ns, verb, name, resURL string) bool {
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

	auth, err := kubernetes.NewForConfig(conn.restConfigOrDie())
	if err != nil {
		log.Warn().Msgf("%s", err)
		return false
	}

	response, err := auth.AuthorizationV1().SelfSubjectAccessReviews().Create(sar)
	if err != nil {
		log.Warn().Msgf("%s", err)
		return false
	}

	return response.Status.Allowed
}
