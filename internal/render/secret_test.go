// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render_test

import (
	"testing"
	"time"

	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSecretRenderPartialMetadata(t *testing.T) {
	secret := &metav1.PartialObjectMetadata{ObjectMeta: metav1.ObjectMeta{
		Namespace:         "default",
		Name:              "database-credentials",
		CreationTimestamp: metav1.NewTime(time.Now().Add(-time.Hour)),
	}}
	row := model1.NewRow(3)

	require.NoError(t, (render.Secret{}).Render(secret, "default", &row))
	assert.Equal(t, "default/database-credentials", row.ID)
	assert.Equal(t, model1.Fields{"default", "database-credentials"}, row.Fields[:2])
	assert.Len(t, row.Fields, 3)
}
