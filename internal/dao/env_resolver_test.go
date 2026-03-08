// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao_test

import (
	"testing"

	"github.com/derailed/k9s/internal/dao"
	"github.com/stretchr/testify/assert"
)

func TestResolveEnvVars(t *testing.T) {
	f := makeFactory()

	uu := map[string]struct {
		input string
		want  string
	}{
		"secret_ref": {
			input: "    DB_PASSWORD:  <set to the key 'token-secret' in secret 'bootstrap-token-abcdef'>  Optional: false",
			want:  "    DB_PASSWORD:  0123456789abcdef",
		},
		"secret_ref_optional_true": {
			input: "    DB_PASSWORD:  <set to the key 'token-secret' in secret 'bootstrap-token-abcdef'>  Optional: true",
			want:  "    DB_PASSWORD:  0123456789abcdef",
		},
		"configmap_ref": {
			input: "    DB_HOST:  <set to the key 'db-host' of config map 'app-config'>  Optional: true",
			want:  "    DB_HOST:  postgres.default.svc",
		},
		"multiple_refs": {
			input: "    DB_PASSWORD:  <set to the key 'token-secret' in secret 'bootstrap-token-abcdef'>  Optional: false\n" +
				"    DB_HOST:  <set to the key 'db-host' of config map 'app-config'>  Optional: true\n" +
				"    DB_PORT:  <set to the key 'db-port' of config map 'app-config'>  Optional: false",
			want: "    DB_PASSWORD:  0123456789abcdef\n" +
				"    DB_HOST:  postgres.default.svc\n" +
				"    DB_PORT:  5432",
		},
		"missing_secret": {
			input: "    FOO:  <set to the key 'bar' in secret 'no-such-secret'>  Optional: false",
			want:  "    FOO:  <error: could not fetch secret 'no-such-secret': resource kube-system/no-such-secret not found>",
		},
		"missing_configmap": {
			input: "    FOO:  <set to the key 'bar' of config map 'no-such-cm'>  Optional: true",
			want:  "    FOO:  <error: could not fetch configmap 'no-such-cm': resource kube-system/no-such-cm not found>",
		},
		"missing_key_in_secret": {
			input: "    FOO:  <set to the key 'no-key' in secret 'bootstrap-token-abcdef'>  Optional: true",
			want:  "    FOO:  <error: key 'no-key' not found in secret 'bootstrap-token-abcdef'>",
		},
		"missing_key_in_configmap": {
			input: "    FOO:  <set to the key 'no-key' of config map 'app-config'>  Optional: false",
			want:  "    FOO:  <error: key 'no-key' not found in configmap 'app-config'>",
		},
		"multiline_value": {
			input: "    PEM_CERT:  <set to the key 'pem-cert' in secret 'pem-secret'>  Optional: false",
			want:  `    PEM_CERT:  -----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhki\n-----END PUBLIC KEY-----\n`,
		},
		"no_refs": {
			input: "    STATIC_VAR:  hello-world",
			want:  "    STATIC_VAR:  hello-world",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			got := dao.ResolveEnvVars(f, u.input, "kube-system/my-pod")
			assert.Equal(t, u.want, got)
		})
	}
}
