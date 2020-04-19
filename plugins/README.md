# K9s community plugins

K9s plugins extend the tool to provide additonal functionality via actions to further help you observe or administer your Kubernetes clusters.

| Plugin-Name     | Description                    | Available on Views | Shortcut | Kubectl plugin, external dependencies     |
|-----------------|--------------------------------|--------------------|----------|-------------------------------------------|
| log_stern.yml   | View resource logs using stern | pods               | Ctrl-l   |                                           |
| log_jq.yml      | View resource logs using jq    | pods               | Ctrl-j   | kubetcl-plugins/kubectl-jq                |
| job_suspend.yml | Suspends a running cronjob     | cronjobs           | Ctrl-s   |                                           |
| dive.yml        | Dive image layers              | containers         | d        | [Dive](https://github.com/wagoodman/dive) |
