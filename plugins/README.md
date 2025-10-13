# K9s community plugins

K9s plugins extend the tool to provide additional functionality via actions to further help you observe or administer
your Kubernetes clusters.

Following is an example of some plugin files in this directory. Other files are not listed in this table.

| Plugin-Name                    | Description                                                                               | Available on Views                  | Shortcut    | Kubectl plugin, external dependencies                                                 |
|--------------------------------|-------------------------------------------------------------------------------------------|-------------------------------------|-------------|---------------------------------------------------------------------------------------|
| ai-incident-investigation.yaml | Run AI investigation on application issues to find the root cause in seconds              | all                                 | Shift-h/o   | [HolmesGPT](https://github.com/robusta-dev/holmesgpt)                                 |
| argocd.yaml                    | Perform argocd operation quickly                                                          | applications                        | Shift-r     | [ArgoCD](https://argo-cd.readthedocs.io/en/stable/getting_started/)                   |
| crd-wizard.yaml                | Clear and intuitive interface for visualizing and exploring CR(D)s                        | applications                        | Shift-w     | [crd-wizard](https://github.com/pehlicd/crd-wizard)                                   |
| debug-container.yaml           | Add [ephemeral debug container](1)<br>([nicolaka/netshoot](2))                            | containers                          | Shift-d     |                                                                                       |
| dive.yaml                      | Dive image layers                                                                         | containers                          | d           | [Dive](https://github.com/wagoodman/dive)                                             |
| dup.yaml                       | Duplicate, edit and Debug resources                                                       | all                                 | Shift-d/e/v | [dup](https://github.com/vash/dup)                                                    |
| external-secrets.yaml          | Refresh external/push-secrets                                                             | externalsecrets/pushsecrets         | Shift-R     | [External Secrets](https://external-secrets.io)                                       |
| get-all-namespace-resources.yaml  | List all namespace resources (using standard kubectl)                                  | all                                 | m           | [kubectl](https://kubernetes.io/docs/tasks/tools/) |
| get-all.yaml                   | get all resources in a namespace                                                          | all                                 | g           | [Krew](https://krew.sigs.k8s.io/), [ketall](https://github.com/corneliusweig/ketall/) |
| helm-diff.yaml                 | Diff with previous revision / current revision                                            | helm/history                        | Shift-D/Q   | [helm-diff](https://github.com/databus23/helm-diff)                                   |
| job-suspend.yaml               | Suspends a running cronjob                                                                | cronjobs                            | Ctrl-s      |                                                                                       |
| k3d-root-shell.yaml            | Root shell to k3d container                                                               | containers                          | Shift-s     | [jq](https://stedolan.github.io/jq/)                                                  |
| keda-toggle.yaml               | Enable/disable [keda](3) ScaledObject autoscaler                                          | scaledobjects                       | Ctrl-N      |                                                                                       |
| kube-metrics.yaml              | Visualize live pod/node metric graphs (Memory/CPU)                                        | pods/nodes                          | m           | [kube-metics](https://github.com/bakito/kube-metrics)                                 |
| log-stern.yaml                 | View resource logs using stern                                                            | pods                                | Ctrl-l      |                                                                                       |
| log-jq.yaml                    | View resource logs using jq                                                               | pods                                | Ctrl-j      | kubectl-plugins/kubectl-jq                                                            |
| log-bunyan.yaml                | View pods, service, deployment logs using bunyan                                          | pods, service, deployment           | Ctrl-l      | [Bunyan](https://www.npmjs.com/package/bunyan)                                        |
| log-full.yaml                  | get full logs from pod/container                                                          | pods/containers                     | Ctrl-l      |                                                                                       |
| pvc-debug-container.yaml       | Add ephemeral debug container with pvc mounted                                            | pods                                | s           | kubectl                                                                               |
| resource-recommendations.yaml  | View recommendations for CPU/Memory requests based on historical data                     | deployments/daemonsets/statefulsets | Shift-k     | [Robusta KRR](https://github.com/robusta-dev/krr)                                     |
| szero.yaml                     | Temporarily scale down/up all deployments, statefulsets, and daemonsets                   | namespaces                          | Shift-d/u   | [szero](https://github.com/jadolg/szero)                                              |
| trace-dns.yaml                 | Trace DNS resolution using Inspektor Gadget (4)                                           | containers/pods/nodes               | Shift-d     |                                                                                       |
| vector-dev-top.yaml            | Run `vector top` in vector.dev container                                                  | pods/container                      | h           | [vector top](https://vector.dev/highlights/2020-12-23-vector-top/)                    |
| start-alpine.yaml              | Starts a deployment for the `alpine:latest` docker image in the current namespace/context | deployments/pods                    | Ctrl-T      |                                                                                       |

[1]: https://kubernetes.io/docs/tasks/debug/debug-application/debug-running-pod/#ephemeral-container

[2]: https://github.com/nicolaka/netshoot

[3]: https://keda.sh/

[4]: https://inspektor-gadget.io/
