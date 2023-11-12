# K9s community plugins

K9s plugins extend the tool to provide additional functionality via actions to further help you observe or administer your Kubernetes clusters.

Following is an example of some of plugin files in this directory. Other files are not listed in this table.

| Plugin-Name        | Description                                                      | Available on Views | Shortcut | Kubectl plugin, external dependencies                                                 |
|--------------------|------------------------------------------------------------------|--------------------|----------|---------------------------------------------------------------------------------------|
| debug-container.yml| Add [ephemeral debug container][1]<br>([nicolaka/netshoot][2]) | containers         | Shift-d  |                                                                                       |
| dive.yml           | Dive image layers                                                | containers         | d        | [Dive](https://github.com/wagoodman/dive)                                             |
| get-all.yml        | get all resources in a namespace                                 | all                | g        | [Krew](https://krew.sigs.k8s.io/), [ketall](https://github.com/corneliusweig/ketall/) |
| job_suspend.yml    | Suspends a running cronjob                                       | cronjobs           | Ctrl-s   |                                                                                       |
| k3d_root_shell.yml | Root shell to k3d container                                      | containers         | Shift-s  | [jq](https://stedolan.github.io/jq/)                                                  |
| log_stern.yml      | View resource logs using stern                                   | pods               | Ctrl-l   |                                                                                       |
| log_jq.yml         | View resource logs using jq                                      | pods               | Ctrl-j   | kubectl-plugins/kubectl-jq                                                            |
| log_full.yml       | get full logs from pod/container                                 | pods/containers    | Ctrl-l   |                                                                                       |

[1]: https://kubernetes.io/docs/tasks/debug/debug-application/debug-running-pod/#ephemeral-container
[2]: https://github.com/nicolaka/netshoot
