package view

import (
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
		<-sigChan
		log.Debug().Msg("Command canceled with signal!")
		cancel()
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

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

const (
	k9sShell           = "k9s-shell"
	k9sShellNS         = "default"
	k9sShellRetryCount = 10
	k9sShellRetryDelay = 500 * time.Millisecond
)

func ssh(a *App, node string) error {
	nukeK9sShell(a.Conn())
	defer nukeK9sShell(a.Conn())
	if err := launchShellPod(a, node); err != nil {
		return err
	}
	shellIn(a, client.FQN(k9sShellNS, k9sShellPodName()), k9sShell)

	return nil
}

func nukeK9sShell(c client.Connection) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	err := c.DialOrDie().CoreV1().Pods(k9sShellNS).Delete(ctx, k9sShellPodName(), metav1.DeleteOptions{})
	if kerrors.IsNotFound(err) {
		return
	}
	if err != nil {
		log.Error().Err(err).Msgf("Fail to delete pod %s", k9sShell)
	}
}

func launchShellPod(a *App, node string) error {
	img := a.Config.K9s.DockerShellImage
	if img == "" {
		img = config.DefaultDockerShellImage
	}
	spec := k9sShellPod(node, img)
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	dial := a.Conn().DialOrDie().CoreV1().Pods(k9sShellNS)
	if _, err := dial.Create(ctx, &spec, metav1.CreateOptions{}); err != nil {
		return err
	}

	for i := 0; i < k9sShellRetryCount; i++ {
		o, err := a.factory.Get("v1/pods", client.FQN(k9sShellNS, k9sShellPodName()), true, labels.Everything())
		if err != nil {
			time.Sleep(k9sShellRetryDelay)
			continue
		}
		var pod v1.Pod
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &pod); err != nil {
			return err
		}
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

func k9sShellPod(node, image string) v1.Pod {
	var grace int64
	var priv bool = true

	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      k9sShellPodName(),
			Namespace: k9sShellNS,
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
			Containers: []v1.Container{
				{
					Name:  k9sShell,
					Image: image,
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      "root-vol",
							MountPath: "/host",
							ReadOnly:  true,
						},
					},
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceCPU:    resource.MustParse("200m"),
							v1.ResourceMemory: resource.MustParse("100Mi"),
						},
					},
					Stdin: true,
					SecurityContext: &v1.SecurityContext{
						Privileged: &priv,
					},
				},
			},
		},
	}
}
