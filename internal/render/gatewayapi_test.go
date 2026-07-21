// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGatewayClassHeader(t *testing.T) {
	gc := GatewayClass{}
	assert.NotNil(t, gc)
	assert.NotNil(t, gc.Base)

	header := gc.Header("")
	assert.Equal(t, 4, len(header))
	assert.Equal(t, "NAME", header[0].Name)
	assert.Equal(t, "CONTROLLER", header[1].Name)
	assert.Equal(t, "AGE", header[2].Name)
	assert.Equal(t, "STATUS", header[3].Name)
}

func TestHTTPRouteHeader(t *testing.T) {
	hr := HTTPRoute{}
	assert.NotNil(t, hr)
	assert.NotNil(t, hr.Base)

	header := hr.Header("")
	assert.Equal(t, 7, len(header))
	assert.Equal(t, "NAMESPACE", header[0].Name)
	assert.Equal(t, "NAME", header[1].Name)
	assert.Equal(t, "HOSTNAMES", header[2].Name)
}

func TestGRPCRouteHeader(t *testing.T) {
	gr := GRPCRoute{}
	assert.NotNil(t, gr)
	assert.NotNil(t, gr.Base)

	header := gr.Header("")
	assert.Equal(t, 7, len(header))
}

func TestTCPRouteHeader(t *testing.T) {
	tr := TCPRoute{}
	assert.NotNil(t, tr)
	assert.NotNil(t, tr.Base)

	header := tr.Header("")
	assert.Equal(t, 6, len(header))
}

func TestUDPRouteHeader(t *testing.T) {
	ur := UDPRoute{}
	assert.NotNil(t, ur)
	assert.NotNil(t, ur.Base)

	header := ur.Header("")
	assert.Equal(t, 6, len(header))
}

func TestTLSRouteHeader(t *testing.T) {
	tr := TLSRoute{}
	assert.NotNil(t, tr)
	assert.NotNil(t, tr.Base)

	header := tr.Header("")
	assert.Equal(t, 7, len(header))
}

func TestReferenceGrantHeader(t *testing.T) {
	rg := ReferenceGrant{}
	assert.NotNil(t, rg)
	assert.NotNil(t, rg.Base)

	header := rg.Header("")
	assert.Equal(t, 5, len(header))
}

func TestBackendTLSPolicyHeader(t *testing.T) {
	btp := BackendTLSPolicy{}
	assert.NotNil(t, btp)
	assert.NotNil(t, btp.Base)

	header := btp.Header("")
	assert.Equal(t, 5, len(header))
}
