# K9s community plugins

K9s plugins extend the tool to provide additional functionality via actions to further help you observe or administer your Kubernetes clusters.

Following is an example of some plugin files in this directory. Other files are not listed in this table.

| Plugin-Name                    | Description                                                                  | Available on Views                  | Shortcut  | Kubectl plugin, external dependencies                                                 |
| ------------------------------ | ---------------------------------------------------------------------------- | ----------------------------------- |-----------| ------------------------------------------------------------------------------------- |
| ai-incident-investigation.yaml | Run AI investigation on application issues to find the root cause in seconds | all                                 | Shift-h/o | [HolmesGPT](https://github.com/robusta-dev/holmesgpt)                                 |
| debug-container.yaml           | Add [ephemeral debug container](1)<br>([nicolaka/netshoot](2))               | containers                          | Shift-d   |                                                                                       |
| dive.yaml                      | Dive image layers                                                            | containers                          | d         | [Dive](https://github.com/wagoodman/dive)                                             |
| get-all.yaml                   | get all resources in a namespace                                             | all                                 | g         | [Krew](https://krew.sigs.k8s.io/), [ketall](https://github.com/corneliusweig/ketall/) |
| helm-diff.yaml                 | Diff with previous revision / current revision                               | helm/history                        | Shift-D/Q | [helm-diff](https://github.com/databus23/helm-diff)                                   |
| job_suspend.yaml               | Suspends a running cronjob                                                   | cronjobs                            | Ctrl-s    |                                                                                       |
| k3d_root_shell.yaml            | Root shell to k3d container                                                  | containers                          | Shift-s   | [jq](https://stedolan.github.io/jq/)                                                  |
| keda-toggle.yaml               | Enable/disable [keda](3) ScaledObject autoscaler                             | scaledobjects                       | Ctrl-N    |                                                                                       |
| log_stern.yaml                 | View resource logs using stern                                               | pods                                | Ctrl-l    |                                                                                       |
| log_jq.yaml                    | View resource logs using jq                                                  | pods                                | Ctrl-j    | kubectl-plugins/kubectl-jq                                                            |
| log_full.yaml                  | get full logs from pod/container                                             | pods/containers                     | Ctrl-l    |                                                                                       |
| resource-recommendations.yaml  | View recommendations for CPU/Memory requests based on historical data        | deployments/daemonsets/statefulsets | Shift-k   | [Robusta KRR](https://github.com/robusta-dev/krr)                                     |
| trace-dns.yaml                 | Trace DNS resolution using Inspektor Gadget (4)                              | containers/pods/nodes               | Shift-d   |                                                                                       |

[1]: https://kubernetes.io/docs/tasks/debug/debug-application/debug-running-pod/#ephemeral-container
[2]: https://github.com/nicolaka/netshoot
[3]: https://keda.sh/
[4]: https://inspektor-gadget.io/
