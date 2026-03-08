// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/derailed/k9s/internal/client"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

// Regexes match kubectl describe output format for env var references.
// See: https://github.com/kubernetes/kubectl/blob/master/pkg/describe/describe.go
var (
	secretKeyRefRE    = regexp.MustCompile(`<set to the key '([^']+)' in secret '([^']+)'>(\s*Optional: (?:true|false))?`)
	configMapKeyRefRE = regexp.MustCompile(`<set to the key '([^']+)' of config map '([^']+)'>(\s*Optional: (?:true|false))?`)
)

type fetchResult struct {
	name        string
	isSecret    bool
	isConfigMap bool
	data        map[string]string
	err         string
}

// sanitizeValue escapes newlines so multi-line values (e.g. PEM certs)
// don't break the one-line-per-env-var describe format.
func sanitizeValue(s string) string {
	return strings.ReplaceAll(s, "\n", `\n`)
}

// ResolveEnvVars resolves secret/configmap env var references in a describe
// output string. Errors are inlined per-reference so partial results are
// always returned.
func ResolveEnvVars(f Factory, desc, path string) string {
	ns, _ := client.Namespaced(path)

	// Collect unique secret and configmap names.
	secretNames := make(map[string]struct{})
	for _, mm := range secretKeyRefRE.FindAllStringSubmatch(desc, -1) {
		if len(mm) >= 3 {
			secretNames[mm[2]] = struct{}{}
		}
	}
	cmNames := make(map[string]struct{})
	for _, mm := range configMapKeyRefRE.FindAllStringSubmatch(desc, -1) {
		if len(mm) >= 3 {
			cmNames[mm[2]] = struct{}{}
		}
	}

	if len(secretNames) == 0 && len(cmNames) == 0 {
		return desc
	}

	// Fetch all secrets and configmaps concurrently.
	var wg sync.WaitGroup
	out := make(chan fetchResult)

	for name := range secretNames {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			out <- fetchSecret(f, ns, name)
		}(name)
	}
	for name := range cmNames {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			out <- fetchConfigMap(f, ns, name)
		}(name)
	}
	go func() {
		wg.Wait()
		close(out)
	}()

	secretCache := make(map[string]fetchResult, len(secretNames))
	cmCache := make(map[string]fetchResult, len(cmNames))
	for res := range out {
		if res.isSecret {
			secretCache[res.name] = res
		}
		if res.isConfigMap {
			cmCache[res.name] = res
		}
	}

	// Replace secret refs using pre-fetched data.
	desc = secretKeyRefRE.ReplaceAllStringFunc(desc, func(match string) string {
		mm := secretKeyRefRE.FindStringSubmatch(match)
		if len(mm) < 3 {
			return match
		}
		key, name := mm[1], mm[2]
		res := secretCache[name]
		if res.err != "" {
			return res.err
		}
		if val, ok := res.data[key]; ok {
			return sanitizeValue(val)
		}
		return fmt.Sprintf("<error: key '%s' not found in secret '%s'>", key, name)
	})

	// Replace configmap refs using pre-fetched data.
	desc = configMapKeyRefRE.ReplaceAllStringFunc(desc, func(match string) string {
		mm := configMapKeyRefRE.FindStringSubmatch(match)
		if len(mm) < 3 {
			return match
		}
		key, name := mm[1], mm[2]
		res := cmCache[name]
		if res.err != "" {
			return res.err
		}
		if val, ok := res.data[key]; ok {
			return sanitizeValue(val)
		}
		return fmt.Sprintf("<error: key '%s' not found in configmap '%s'>", key, name)
	})

	return desc
}

func fetchSecret(f Factory, ns, name string) fetchResult {
	fqn := FQN(ns, name)
	res := fetchResult{name: name, isSecret: true}

	o, err := f.Get(client.SecGVR, fqn, true, labels.Everything())
	if err != nil {
		res.err = fmt.Sprintf("<error: could not fetch secret '%s': %s>", name, err)
		return res
	}
	data, err := ExtractSecrets(o)
	if err != nil {
		res.err = fmt.Sprintf("<error: could not decode secret '%s': %s>", name, err)
		return res
	}
	res.data = data

	return res
}

func fetchConfigMap(f Factory, ns, name string) fetchResult {
	fqn := FQN(ns, name)
	res := fetchResult{name: name, isConfigMap: true}

	o, err := f.Get(client.CmGVR, fqn, true, labels.Everything())
	if err != nil {
		res.err = fmt.Sprintf("<error: could not fetch configmap '%s': %s>", name, err)
		return res
	}
	u, ok := o.(*unstructured.Unstructured)
	if !ok {
		res.err = fmt.Sprintf("<error: could not decode configmap '%s': unexpected type %T>", name, o)
		return res
	}
	var cm v1.ConfigMap
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &cm); err != nil {
		res.err = fmt.Sprintf("<error: could not decode configmap '%s': %s>", name, err)
		return res
	}
	res.data = cm.Data

	return res
}
