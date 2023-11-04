import * as pulumi from "@pulumi/pulumi";
import * as aws from "@pulumi/aws";
import * as awsx from "@pulumi/awsx";

const provider = new aws.Provider("localstack", {
  region: "us-east-1",
  accessKey: "fake",
  secretKey: "fake",
  skipCredentialsValidation: true,
  skipMetadataApiCheck: true,
  skipRequestingAccountId: true,
  s3UsePathStyle: true,

  endpoints: [
    {
      s3: "http://localhost:4566",
      sqs: "http://localhost:4566",
    },
  ],
});

const opts: pulumi.CustomResourceOptions = { provider };
const ingest = new aws.s3.Bucket("ingest", { bucket: "ingest" }, opts);
const query = new aws.s3.Bucket("query", { bucket: "query" }, opts);

function getSQSPolicy(
  queueName: string,
  arn: pulumi.Output<string>
): pulumi.Output<aws.iam.PolicyDocument> {
  return arn.apply((arn) => {
    const doc: aws.iam.PolicyDocument = {
      Version: "2012-10-17",
      Statement: [
        {
          Effect: "Allow",
          Principal: "*",
          Action: "sqs:SendMessage",
          Resource: `arn:aws:sqs:*:*:${queueName}`,
          Condition: {
            ArnEquals: {
              "aws:SourceArn": arn,
            },
          },
        },
      ],
    };

    return doc;
  });
}

const queue = new aws.sqs.Queue(
  "ingest-queue",
  {
    name: "ingest-queue",
    policy: getSQSPolicy("ingest-queue", ingest.arn).apply(JSON.stringify),
  },
  opts
);

const bucketNotification = new aws.s3.BucketNotification(
  "ingest-bucket-notification",
  {
    bucket: ingest.bucket,
    queues: [
      {
        events: ["s3:ObjectCreated:*"],
        queueArn: queue.arn,
      },
    ],
  },
  opts
);

export const ingestBucket = ingest.bucket;
export const queryBucket = query.bucket;
export const queueUrl = queue.id;
