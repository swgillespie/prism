#!/bin/bash

set -euo pipefail

echo "> Booting Localstack"
overmind start -D --procfile Procfile.infra -l localstack

echo "> Bootstrapping Localstack"
sleep 5
awslocal s3api create-bucket --bucket ingest --region us-east-1
awslocal s3api create-bucket --bucket query --region us-east-1

echo "> Initializing CockroachDB"
psql -Atx "postgresql://$COCKROACHDB_USER:$COCKROACHDB_PASSWORD@$COCKROACHDB_URL/$COCKROACHDB_DATABASE?sslmode=verify-full" -f ./.scratch/meta.sql 

echo "> Putting benchmark data into Localstack"
awslocal s3 cp ./.scratch/testdir s3://query/public/hits --recursive