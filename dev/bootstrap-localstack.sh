#!/bin/bash

set -euo pipefail

awslocal s3api create-bucket --bucket ingest --region us-east-1
awslocal s3api create-bucket --bucket query --region us-east-1