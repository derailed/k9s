// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

// BOZO!! Revamp with latest

// import (
// 	"bytes"
// 	"context"
// 	"encoding/json"
// 	"errors"
// 	"fmt"
// 	"io"
// 	"net/http"
// 	"net/url"
// 	"os"
// 	"path"
// 	"strings"
// 	"time"

// 	"github.com/derailed/k9s/internal/client"
// 	"github.com/derailed/k9s/internal/render"
// 	"github.com/openfaas/faas-cli/proxy"
// 	"github.com/openfaas/faas/gateway/requests"
// 	"github.com/rs/zerolog/log"
// 	"k8s.io/apimachinery/pkg/runtime"
// 	"sigs.k8s.io/yaml"
// )

// const (
// 	oFaasGatewayEnv  = "OPENFAAS_GATEWAY"
// 	oFaasJWTTokenEnv = "OPENFAAS_JWT_TOKEN"
// 	oFaasTLSInsecure = "OPENFAAS_TLS_INSECURE"
// )

// var (
// 	_ Accessor  = (*OpenFaas)(nil)
// 	_ Nuker     = (*OpenFaas)(nil)
// 	_ Describer = (*OpenFaas)(nil)
// )

// // OpenFaas represents a faas gateway connection.
// type OpenFaas struct {
// 	NonResource
// }

// // IsOpenFaasEnabled returns true if a gateway url is set in the environment.
// func IsOpenFaasEnabled() bool {
// 	return os.Getenv(oFaasGatewayEnv) != ""
// }

// func getOpenFAASFlags() (string, string, bool) {
// 	gw, token := os.Getenv(oFaasGatewayEnv), os.Getenv(oFaasJWTTokenEnv)
// 	tlsInsecure := false
// 	if os.Getenv(oFaasTLSInsecure) == "true" {
// 		tlsInsecure = true
// 	}

// 	return gw, token, tlsInsecure
// }

// // Get returns a function by name.
// func (f *OpenFaas) Get(ctx context.Context, path string) (runtime.Object, error) {
// 	ns, n := client.Namespaced(path)

// 	oo, err := f.List(ctx, ns)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var found runtime.Object
// 	for _, o := range oo {
// 		r, ok := o.(render.OpenFaasRes)
// 		if !ok {
// 			continue
// 		}
// 		if r.Function.Name == n {
// 			found = o
// 			break
// 		}
// 	}

// 	if found == nil {
// 		return nil, fmt.Errorf("unable to locate function %q", path)
// 	}

// 	return found, nil
// }

// // List returns a collection of functions.
// func (f *OpenFaas) List(_ context.Context, ns string) ([]runtime.Object, error) {
// 	if !IsOpenFaasEnabled() {
// 		return nil, errors.New("OpenFAAS is not enabled on this cluster")
// 	}

// 	gw, token, tls := getOpenFAASFlags()
// 	ff, err := proxy.ListFunctionsToken(gw, tls, token, ns)
// 	if err != nil {
// 		return nil, err
// 	}

// 	oo := make([]runtime.Object, 0, len(ff))
// 	for _, f := range ff {
// 		oo = append(oo, render.OpenFaasRes{Function: f})
// 	}

// 	return oo, nil
// }

// // Delete removes a function.
// func (f *OpenFaas) Delete(path string, _, _ bool) error {
// 	gw, token, tls := getOpenFAASFlags()
// 	ns, n := client.Namespaced(path)

// 	// BOZO!! openfaas spews to stdout. Not good for us...
// 	return deleteFunctionToken(gw, n, tls, token, ns)
// }

// // ToYAML dumps a function to yaml.
// func (f *OpenFaas) ToYAML(path string, _ bool) (string, error) {
// 	return f.Describe(path)
// }

// // Describe describes a function.
// func (f *OpenFaas) Describe(path string) (string, error) {
// 	o, err := f.Get(context.Background(), path)
// 	if err != nil {
// 		return "", err
// 	}

// 	fn, ok := o.(render.OpenFaasRes)
// 	if !ok {
// 		return "", fmt.Errorf("expecting OpenFaasRes but got %T", o)
// 	}

// 	raw, err := json.Marshal(fn)
// 	if err != nil {
// 		return "", err
// 	}

// 	bytes, err := yaml.JSONToYAML(raw)
// 	if err != nil {
// 		return "", err
// 	}

// 	return string(bytes), nil
// }

// // BOZO!! Meow! openfaas fn prints to stdout have to dup ;(.
// func deleteFunctionToken(gateway string, functionName string, tlsInsecure bool, token string, namespace string) error {
// 	defaultCommandTimeout := 60 * time.Second

// 	gateway = strings.TrimRight(gateway, "/")
// 	delReq := requests.DeleteFunctionRequest{FunctionName: functionName}
// 	reqBytes, _ := json.Marshal(&delReq)
// 	reader := bytes.NewReader(reqBytes)

// 	c := proxy.MakeHTTPClient(&defaultCommandTimeout, tlsInsecure)

// 	deleteEndpoint, err := createSystemEndpoint(gateway, namespace)
// 	if err != nil {
// 		return err
// 	}

// 	req, err := http.NewRequestWithContext(context.Background(), "DELETE", deleteEndpoint, reader)
// 	if err != nil {
// 		return err
// 	}
// 	req.Header.Set("Content-Type", "application/json")

// 	if len(token) > 0 {
// 		proxy.SetToken(req, token)
// 	} else {
// 		proxy.SetAuth(req, gateway)
// 	}

// 	delRes, delErr := c.Do(req)
// 	if delErr != nil {
// 		return delErr
// 	}

// 	if delRes.Body != nil {
// 		defer func() {
// 			if err := delRes.Body.Close(); err != nil {
// 				log.Error().Err(err).Msgf("closing delete-gtw body")
// 			}
// 		}()
// 	}

// 	switch delRes.StatusCode {
// 	case http.StatusOK, http.StatusCreated, http.StatusAccepted:
// 		return nil
// 	case http.StatusNotFound:
// 		return fmt.Errorf("no function named %s found", functionName)
// 	case http.StatusUnauthorized:
// 		return fmt.Errorf("unauthorized access, run \"faas-cli login\" to setup authentication for this server")
// 	default:
// 		bytesOut, err := io.ReadAll(delRes.Body)
// 		if err != nil {
// 			return err
// 		}
// 		return fmt.Errorf("server returned unexpected status code %d %s", delRes.StatusCode, string(bytesOut))
// 	}
// }

// func createSystemEndpoint(gateway, namespace string) (string, error) {
// 	const systemPath = "/system/functions"

// 	gatewayURL, err := url.Parse(gateway)
// 	if err != nil {
// 		return "", fmt.Errorf("invalid gateway URL: %w", err)
// 	}
// 	gatewayURL.Path = path.Join(gatewayURL.Path, systemPath)
// 	if len(namespace) > 0 {
// 		q := gatewayURL.Query()
// 		q.Set("namespace", namespace)
// 		gatewayURL.RawQuery = q.Encode()
// 	}
// 	return gatewayURL.String(), nil
// }
