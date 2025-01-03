// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/derailed/k9s/internal/render"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/fatih/color"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	shellCheck   = `command -v bash >/dev/null && exec bash || exec sh`
	bannerFmt    = "<<K9s-Shell>> Pod: %s | Container: %s \n"
	outputPrefix = "[output]"
)

var editorEnvVars = []string{"K9S_EDITOR", "KUBE_EDITOR", "EDITOR"}

type shellOpts struct {
	clear, background bool
	pipes             []string
	binary            string
	banner            string
	args              []string
}

func (s shellOpts) String() string {
	return fmt.Sprintf("%s %s", s.binary, strings.Join(s.args, " "))
}

func runK(a *App, opts shellOpts) error {
	bin, err := exec.LookPath("kubectl")
	if errors.Is(err, exec.ErrDot) {
		return fmt.Errorf("kubectl command must not be in the current working directory: %w", err)
	}
	if err != nil {
		return fmt.Errorf("kubectl command is not in your path: %w", err)
	}
	args := []string{opts.args[0]}
	if u, err := a.Conn().Config().ImpersonateUser(); err == nil {
		args = append(args, "--as", u)
	}
	if g, err := a.Conn().Config().ImpersonateGroups(); err == nil {
		args = append(args, "--as-group", g)
	}
	if isInsecure := a.Conn().Config().Flags().Insecure; isInsecure != nil && *isInsecure {
		args = append(args, "--insecure-skip-tls-verify")
	}
	args = append(args, "--context", a.Config.K9s.ActiveContextName())
	if cfg := a.Conn().Config().Flags().KubeConfig; cfg != nil && *cfg != "" {
		args = append(args, "--kubeconfig", *cfg)
	}
	if len(args) > 0 {
		opts.args = append(args, opts.args[1:]...)
	}
	opts.binary = bin

	suspended, errChan, stChan := run(a, opts)
	if !suspended {
		return fmt.Errorf("unable to run command")
	}
	for v := range stChan {
		log.Debug().Msgf("  - %s", v)
	}
	var errs error
	for e := range errChan {
		errs = errors.Join(errs, e)
	}

	return errs
}

func run(a *App, opts shellOpts) (bool, chan error, chan string) {
	errChan := make(chan error, 1)
	statusChan := make(chan string, 1)

	if opts.background {
		if err := execute(opts, statusChan); err != nil {
			errChan <- err
			a.Flash().Errf("Exec failed %q: %s", opts, err)
		}
		close(errChan)
		return true, errChan, statusChan
	}

	a.Halt()
	defer a.Resume()

	return a.Suspend(func() {
		if err := execute(opts, statusChan); err != nil {
			errChan <- err
			a.Flash().Errf("Exec failed %q: %s", opts, err)
		}
		close(errChan)
	}), errChan, statusChan
}

func edit(a *App, opts shellOpts) bool {
	var (
		bin string
		err error
	)
	for _, e := range editorEnvVars {
		env := os.Getenv(e)
		if env == "" {
			continue
		}

		// There may be situations where the user sets the editor as the binary
		// followed by some arguments (e.g. "code -w" to make it work with vscode)
		//
		// In such cases, the actual binary is only the first token
		envTokens := strings.Split(env, " ")

		if bin, err = exec.LookPath(envTokens[0]); err == nil {
			// Make sure the path is at the end (this allows running editors
			// with custom options)
			if len(envTokens) > 1 {
				originalArgs := opts.args
				opts.args = envTokens[1:]
				opts.args = append(opts.args, originalArgs...)
			}

			break
		}
	}
	if bin == "" {
		a.Flash().Errf("You must set at least one of those env vars: %s", strings.Join(editorEnvVars, "|"))
		return false
	}
	opts.binary, opts.background = bin, false

	suspended, errChan, _ := run(a, opts)
	if !suspended {
		a.Flash().Errf("edit command failed")
	}
	status := true
	for e := range errChan {
		a.Flash().Err(e)
		status = false
	}

	return status
}

func execute(opts shellOpts, statusChan chan<- string) error {
	if opts.clear {
		clearScreen()
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		if !opts.background {
			cancel()
			clearScreen()
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func(cancel context.CancelFunc) {
		defer log.Debug().Msgf("SIGNAL_GOR - BAILED!!")
		select {
		case sig := <-sigChan:
			log.Debug().Msgf("Command canceled with signal! %#v", sig)
			cancel()
		case <-ctx.Done():
			log.Debug().Msgf("SIGNAL Context CANCELED!")
		}
	}(cancel)

	cmds := make([]*exec.Cmd, 0, 1)
	cmd := exec.CommandContext(ctx, opts.binary, opts.args...)
	log.Debug().Msgf("RUNNING> %s", opts)

	if env := os.Getenv("K9S_EDITOR"); env != "" {
		// There may be situations where the user sets the editor as the binary
		// followed by some arguments (e.g. "code -w" to make it work with vscode)
		//
		// In such cases, the actual binary is only the first token
		binTokens := strings.Split(env, " ")

		if bin, err := exec.LookPath(binTokens[0]); err == nil {
			binTokens[0] = bin
			cmd.Env = append(os.Environ(), fmt.Sprintf("KUBE_EDITOR=%s", strings.Join(binTokens, " ")))
		}
	}

	cmds = append(cmds, cmd)

	for _, p := range opts.pipes {
		tokens := strings.Split(p, " ")
		if len(tokens) < 2 {
			continue
		}
		cmd := exec.CommandContext(ctx, tokens[0], tokens[1:]...)
		log.Debug().Msgf("\t| %s", cmd)
		cmds = append(cmds, cmd)
	}

	var o, e bytes.Buffer
	err := pipe(ctx, opts, statusChan, &o, &e, cmds...)
	if err != nil {
		log.Err(err).Msgf("Command failed")
		return errors.Join(err, fmt.Errorf("%s", e.String()))
	}

	return nil
}

func runKu(a *App, opts shellOpts) (string, error) {
	bin, err := exec.LookPath("kubectl")
	if errors.Is(err, exec.ErrDot) {
		log.Error().Err(err).Msgf("kubectl command must not be in the current working directory")
		return "", err
	}
	if err != nil {
		log.Error().Err(err).Msgf("kubectl command is not in your path")
		return "", err
	}
	var args []string
	if u, err := a.Conn().Config().ImpersonateUser(); err == nil {
		args = append(args, "--as", u)
	}
	if g, err := a.Conn().Config().ImpersonateGroups(); err == nil {
		args = append(args, "--as-group", g)
	}
	args = append(args, "--context", a.Config.K9s.ActiveContextName())
	if cfg := a.Conn().Config().Flags().KubeConfig; cfg != nil && *cfg != "" {
		args = append(args, "--kubeconfig", *cfg)
	}
	if len(args) > 0 {
		opts.args = append(args, opts.args...)
	}
	opts.binary, opts.background = bin, false

	return oneShoot(opts)
}

func oneShoot(opts shellOpts) (string, error) {
	if opts.clear {
		clearScreen()
	}

	log.Debug().Msgf("Running command> %s %s", opts.binary, strings.Join(opts.args, " "))
	cmd := exec.Command(opts.binary, opts.args...)

	var err error
	buff := bytes.NewBufferString("")
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, buff, buff
	_, _ = cmd.Stdout.Write([]byte(opts.banner))
	err = cmd.Run()

	return strings.Trim(buff.String(), "\n"), err
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

const (
	k9sShell           = "k9s-shell"
	k9sShellRetryCount = 50
	k9sShellRetryDelay = 2 * time.Second
)

func launchNodeShell(v model.Igniter, a *App, node string) {
	if err := nukeK9sShell(a); err != nil {
		a.Flash().Errf("Cleaning node shell failed: %s", err)
		return
	}

	msg := fmt.Sprintf("Launching node shell on %s...", node)
	dialog.ShowPrompt(a.Styles.Dialog(), a.Content.Pages, "Launching", msg, func(ctx context.Context) {
		err := launchShellPod(ctx, a, node)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				a.Flash().Errf("Launching node shell failed: %s", err)
			}
			return
		}

		go launchPodShell(v, a)
	}, func() {
		if err := nukeK9sShell(a); err != nil {
			a.Flash().Errf("Cleaning node shell failed: %s", err)
			return
		}
	})
}

func launchPodShell(v model.Igniter, a *App) {
	defer func() {
		if err := nukeK9sShell(a); err != nil {
			a.Flash().Errf("Launching node shell failed: %s", err)
			return
		}
	}()

	v.Stop()
	defer v.Start()

	ns := a.Config.K9s.ShellPod.Namespace
	if err := sshIn(a, client.FQN(ns, k9sShellPodName()), k9sShell); err != nil {
		a.Flash().Errf("Launching node shell failed: %s", err)
	}
}

func sshIn(a *App, fqn, co string) error {
	cfg := a.Config.K9s.ShellPod
	os, err := getPodOS(a.factory, fqn)
	if err != nil {
		return fmt.Errorf("os detect failed: %w", err)
	}

	args := buildShellArgs("exec", fqn, co, a.Conn().Config().Flags().KubeConfig)
	args = append(args, "--")
	if len(cfg.Command) > 0 {
		args = append(args, cfg.Command...)
		args = append(args, cfg.Args...)
	} else {
		if os == windowsOS {
			args = append(args, "--", powerShell)
		}
		args = append(args, "sh", "-c", shellCheck)
	}
	log.Debug().Msgf("ARGS %#v", args)

	c := color.New(color.BgGreen).Add(color.FgBlack).Add(color.Bold)
	err = runK(a, shellOpts{clear: true, banner: c.Sprintf(bannerFmt, fqn, co), args: args})
	if err != nil {
		return fmt.Errorf("shell exec failed: %w", err)
	}

	return nil
}

func nukeK9sShell(a *App) error {
	ct, err := a.Config.K9s.ActiveContext()
	if err != nil {
		return err
	}
	if !ct.FeatureGates.NodeShell {
		return nil
	}

	ns := a.Config.K9s.ShellPod.Namespace
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	dial, err := a.Conn().Dial()
	if err != nil {
		return err
	}

	err = dial.CoreV1().Pods(ns).Delete(ctx, k9sShellPodName(), metav1.DeleteOptions{})
	if kerrors.IsNotFound(err) {
		return nil
	}

	return err
}

func launchShellPod(ctx context.Context, a *App, node string) error {
	var (
		spo  = a.Config.K9s.ShellPod
		spec = k9sShellPod(node, spo)
	)

	dial, err := a.Conn().Dial()
	if err != nil {
		return err
	}

	conn := dial.CoreV1().Pods(spo.Namespace)
	if _, err = conn.Create(ctx, spec, metav1.CreateOptions{}); err != nil {
		return err
	}

	for i := 0; i < k9sShellRetryCount; i++ {
		o, err := a.factory.Get("v1/pods", client.FQN(spo.Namespace, k9sShellPodName()), true, labels.Everything())
		if err != nil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(k9sShellRetryDelay):
				continue
			}
		}

		var pod v1.Pod
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &pod); err != nil {
			return err
		}
		log.Debug().Msgf("Checking shell pod [%d] %v", i, pod.Status.Phase)
		if pod.Status.Phase == v1.PodRunning {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(k9sShellRetryDelay):
		}
	}

	return fmt.Errorf("unable to launch shell pod on node %s", node)
}

func k9sShellPodName() string {
	return fmt.Sprintf("%s-%d", k9sShell, os.Getpid())
}

func k9sShellPod(node string, cfg config.ShellPod) *v1.Pod {
	var grace int64
	var priv bool = true

	log.Debug().Msgf("Shell Config %#v", cfg)
	c := v1.Container{
		Name:            k9sShell,
		Image:           cfg.Image,
		ImagePullPolicy: cfg.ImagePullPolicy,
		VolumeMounts: []v1.VolumeMount{
			{
				Name:      "root-vol",
				MountPath: "/host",
				ReadOnly:  true,
			},
		},
		Resources: asResource(cfg.Limits),
		Stdin:     true,
		TTY:       cfg.TTY,
		SecurityContext: &v1.SecurityContext{
			Privileged: &priv,
		},
	}
	if len(cfg.Command) != 0 {
		c.Command = cfg.Command
	}
	if len(cfg.Args) > 0 {
		c.Args = cfg.Args
	}

	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      k9sShellPodName(),
			Namespace: cfg.Namespace,
			Labels:    cfg.Labels,
		},
		Spec: v1.PodSpec{
			NodeName:                      node,
			RestartPolicy:                 v1.RestartPolicyNever,
			HostPID:                       true,
			HostNetwork:                   true,
			ImagePullSecrets:              cfg.ImagePullSecrets,
			TerminationGracePeriodSeconds: &grace,
			Volumes: []v1.Volume{
				{
					Name: "root-vol",
					VolumeSource: v1.VolumeSource{
						HostPath: &v1.HostPathVolumeSource{
							Path: "/",
						},
					},
				},
			},
			Containers: []v1.Container{c},
			Tolerations: []v1.Toleration{
				{
					Operator: v1.TolerationOperator("Exists"),
				},
			},
		},
	}
}

func asResource(r config.Limits) v1.ResourceRequirements {
	return v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse(r[v1.ResourceCPU]),
			v1.ResourceMemory: resource.MustParse(r[v1.ResourceMemory]),
		},
	}
}

func pipe(_ context.Context, opts shellOpts, statusChan chan<- string, w, e *bytes.Buffer, cmds ...*exec.Cmd) error {
	if len(cmds) == 0 {
		return nil
	}

	if len(cmds) == 1 {
		cmd := cmds[0]
		if opts.background {
			go func() {
				cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, w, e
				if err := cmd.Run(); err != nil {
					log.Error().Err(err).Msgf("Command failed: %s", err)
				} else {
					for _, l := range strings.Split(w.String(), "\n") {
						if l != "" {
							statusChan <- fmt.Sprintf("%s %s", outputPrefix, l)
						}
					}
					statusChan <- fmt.Sprintf("Command completed successfully: %q", render.Truncate(cmd.String(), 20))
					log.Info().Msgf("Command completed successfully: %q", cmd.String())
				}
				close(statusChan)
			}()
			return nil
		}
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
		_, _ = cmd.Stdout.Write([]byte(opts.banner))

		log.Debug().Msgf("Running Start")
		err := cmd.Run()
		log.Debug().Msgf("Running Done: %v", err)
		if err == nil {
			statusChan <- fmt.Sprintf("Command completed successfully: %q", cmd.String())
		}
		close(statusChan)

		return err
	}

	last := len(cmds) - 1
	for i := 0; i < len(cmds); i++ {
		cmds[i].Stderr = os.Stderr
		if i+1 < len(cmds) {
			r, w := io.Pipe()
			cmds[i].Stdout, cmds[i+1].Stdin = w, r
		}
	}
	cmds[last].Stdout = os.Stdout

	for _, cmd := range cmds {
		log.Debug().Msgf("Starting CMD %s", cmd)
		if err := cmd.Start(); err != nil {
			return err
		}
	}

	return cmds[len(cmds)-1].Wait()
}
