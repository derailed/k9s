package view

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
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
	shellCheck = `command -v bash >/dev/null && exec bash || exec sh`
	bannerFmt  = "<<K9s-Shell>> Pod: %s | Container: %s \n"
)

type shellOpts struct {
	clear, background bool
	binary            string
	banner            string
	args              []string
}

func runK(a *App, opts shellOpts) bool {
	bin, err := exec.LookPath("kubectl")
	if err != nil {
		log.Error().Err(err).Msgf("kubectl command is not in your path")
		return false
	}
	var args []string
	if u, err := a.Conn().Config().ImpersonateUser(); err == nil {
		args = append(args, "--as", u)
	}
	if g, err := a.Conn().Config().ImpersonateGroups(); err == nil {
		args = append(args, "--as-group", g)
	}
	if isInsecure := a.Conn().Config().Flags().Insecure; isInsecure != nil && *isInsecure {
		args = append(args, "--insecure-skip-tls-verify")
	}
	args = append(args, "--context", a.Config.K9s.CurrentContext)
	if cfg := a.Conn().Config().Flags().KubeConfig; cfg != nil && *cfg != "" {
		args = append(args, "--kubeconfig", *cfg)
	}
	if len(args) > 0 {
		opts.args = append(args, opts.args...)
	}
	opts.binary, opts.background = bin, false

	return run(a, opts)
}

func run(a *App, opts shellOpts) bool {
	a.Halt()
	defer a.Resume()

	return a.Suspend(func() {
		if err := execute(opts); err != nil {
			a.Flash().Errf("Command exited: %v", err)
		}
	})
}

func edit(a *App, opts shellOpts) bool {
	bin, err := exec.LookPath(os.Getenv("K9S_EDITOR"))
	if err != nil {
		bin, err = exec.LookPath(os.Getenv("EDITOR"))
		if err != nil {
			log.Error().Err(err).Msgf("K9S_EDITOR|EDITOR not set")
			return false
		}
	}
	opts.binary, opts.background = bin, false

	return run(a, opts)
}

func execute(opts shellOpts) error {
	if opts.clear {
		clearScreen()
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		clearScreen()
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case <-sigChan:
			log.Debug().Msg("Command canceled with signal!")
			cancel()
		case <-ctx.Done():
			return
		}
	}()

	log.Debug().Msgf("Running command> %s %s", opts.binary, strings.Join(opts.args, " "))
	cmd := exec.Command(opts.binary, opts.args...)

	var err error
	if opts.background {
		err = cmd.Start()
	} else {
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
		_, _ = cmd.Stdout.Write([]byte(opts.banner))
		err = cmd.Run()
	}

	select {
	case <-ctx.Done():
		return errors.New("canceled by operator")
	default:
		return err
	}
}

func runKu(a *App, opts shellOpts) (string, error) {
	bin, err := exec.LookPath("kubectl")
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
	args = append(args, "--context", a.Config.K9s.CurrentContext)
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
	k9sShellRetryCount = 10
	k9sShellRetryDelay = 1 * time.Second
)

func ssh(a *App, node string) error {
	if err := nukeK9sShell(a); err != nil {
		return err
	}
	defer func() {
		if err := nukeK9sShell(a); err != nil {
			log.Error().Err(err).Msgf("nuking k9s shell pod")
		}
	}()
	if err := launchShellPod(a, node); err != nil {
		return err
	}
	ns := a.Config.K9s.ActiveCluster().ShellPod.Namespace
	sshIn(a, client.FQN(ns, k9sShellPodName()), k9sShell)

	return nil
}

func sshIn(a *App, fqn, co string) {
	cfg := a.Config.K9s.ActiveCluster().ShellPod
	os, err := getPodOS(a.factory, fqn)
	if err != nil {
		log.Warn().Err(err).Msgf("os detect failed")
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
	if !runK(a, shellOpts{clear: true, banner: c.Sprintf(bannerFmt, fqn, co), args: args}) {
		a.Flash().Err(errors.New("Shell exec failed"))
	}
}

func nukeK9sShell(a *App) error {
	cl := a.Config.K9s.CurrentCluster
	if !a.Config.K9s.Clusters[cl].FeatureGates.NodeShell {
		return nil
	}

	ns := a.Config.K9s.ActiveCluster().ShellPod.Namespace
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

func launchShellPod(a *App, node string) error {
	a.Flash().Infof("Launching node shell on %s...", node)
	ns := a.Config.K9s.ActiveCluster().ShellPod.Namespace
	spec := k9sShellPod(node, a.Config.K9s.ActiveCluster().ShellPod)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dial, err := a.Conn().Dial()
	if err != nil {
		return err
	}
	conn := dial.CoreV1().Pods(ns)
	if _, err := conn.Create(ctx, &spec, metav1.CreateOptions{}); err != nil {
		return err
	}

	for i := 0; i < k9sShellRetryCount; i++ {
		o, err := a.factory.Get("v1/pods", client.FQN(ns, k9sShellPodName()), true, labels.Everything())
		if err != nil {
			time.Sleep(k9sShellRetryDelay)
			continue
		}
		var pod v1.Pod
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &pod); err != nil {
			return err
		}
		log.Debug().Msgf("Checking shell pod [%d] %v", i, pod.Status.Phase)
		if pod.Status.Phase == v1.PodRunning {
			return nil
		}
		time.Sleep(k9sShellRetryDelay)
	}

	return fmt.Errorf("Unable to launch shell pod on node %s", node)
}

func k9sShellPodName() string {
	return fmt.Sprintf("%s-%d", k9sShell, os.Getpid())
}

func k9sShellPod(node string, cfg *config.ShellPod) v1.Pod {
	var grace int64
	var priv bool = true

	log.Debug().Msgf("Shell Config %#v", cfg)
	c := v1.Container{
		Name:  k9sShell,
		Image: cfg.Image,
		VolumeMounts: []v1.VolumeMount{
			{
				Name:      "root-vol",
				MountPath: "/host",
				ReadOnly:  true,
			},
		},
		Resources: asResource(cfg.Limits),
		Stdin:     true,
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

	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      k9sShellPodName(),
			Namespace: cfg.Namespace,
		},
		Spec: v1.PodSpec{
			NodeName:                      node,
			RestartPolicy:                 v1.RestartPolicyNever,
			HostPID:                       true,
			HostNetwork:                   true,
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
