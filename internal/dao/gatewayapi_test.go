// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGatewayClass(t *testing.T) {
	gc := &GatewayClass{}
	assert.NotNil(t, gc)
	assert.NotNil(t, &gc.Resource)
}

func TestHTTPRoute(t *testing.T) {
	hr := &HTTPRoute{}
	assert.NotNil(t, hr)
	assert.NotNil(t, &hr.Resource)
}

func TestGRPCRoute(t *testing.T) {
	gr := &GRPCRoute{}
	assert.NotNil(t, gr)
	assert.NotNil(t, &gr.Resource)
}

func TestTCPRoute(t *testing.T) {
	tr := &TCPRoute{}
	assert.NotNil(t, tr)
	assert.NotNil(t, &tr.Resource)
}

func TestUDPRoute(t *testing.T) {
	ur := &UDPRoute{}
	assert.NotNil(t, ur)
	assert.NotNil(t, &ur.Resource)
}

func TestTLSRoute(t *testing.T) {
	tr := &TLSRoute{}
	assert.NotNil(t, tr)
	assert.NotNil(t, &tr.Resource)
}

func TestReferenceGrant(t *testing.T) {
	rg := &ReferenceGrant{}
	assert.NotNil(t, rg)
	assert.NotNil(t, &rg.Resource)
}

func TestBackendTLSPolicy(t *testing.T) {
	btp := &BackendTLSPolicy{}
	assert.NotNil(t, btp)
	assert.NotNil(t, &btp.Resource)
}
