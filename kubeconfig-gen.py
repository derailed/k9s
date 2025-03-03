#!/usr/local/bin/python3
import json
import subprocess
import argparse

def main(args):
    cmd = "mkdir -p ~/.kube"
    subprocess.run(cmd, shell=True)
    cmd = f"linode-cli lke clusters-list --json"
    output = subprocess.check_output(cmd, shell=True).decode('utf-8')
    result = json.loads(output)
    cluster_id = ""
    for cluster in result:
        if cluster['label'] == args.cluster:
            cluster_id = cluster['id']
            break
    cmd = f"linode-cli lke kubeconfig-view {cluster_id} --json"
    output = subprocess.check_output(cmd, shell=True).decode('utf-8')
    result = json.loads(output)
    cmd = f"echo '{result[0]['kubeconfig']}' | base64 -d > ~/.kube/config && chmod 600 ~/.kube/config"
    subprocess.run(cmd, shell=True)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description='Generate kubeconfig for Linode Kubernetes Engine')
    parser.add_argument('--cluster', type=str, help='Cluster name')

    args = parser.parse_args()
    main(args)
    print("kubeconfig generated successfully")
