#!/bin/bash

kubectl logs -f $1 -n $2 --context $3 | jq -rR '. as $raw | try (fromjson | .message) catch ("\u001b[31m" + $raw + "\u001b[0m")'
