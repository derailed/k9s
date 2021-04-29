# K9s community plugins

K9s plugins extend the tool to provide additional functionality via actions to further help you observe or administer your Kubernetes clusters.

| Plugin-Name     | Description                      | Available on Views | Shortcut | Kubectl plugin, external dependencies                                                 |
|-----------------|----------------------------------|--------------------|----------|---------------------------------------------------------------------------------------|
| log_stern.yml   | View resource logs using stern   | pods               | Ctrl-l   |                                                                                       |
| log_jq.yml      | View resource logs using jq      | pods               | Ctrl-j   | kubectl-plugins/kubectl-jq                                                            |
| job_suspend.yml | Suspends a running cronjob       | cronjobs           | Ctrl-s   |                                                                                       |
| dive.yml        | Dive image layers                | containers         | d        | [Dive](https://github.com/wagoodman/dive)                                             |
| get-all.yml     | get all resources in a namespace | all                | g        | [Krew](https://krew.sigs.k8s.io/), [ketall](https://github.com/corneliusweig/ketall/) |
