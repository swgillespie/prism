import * as aws from "@pulumi/aws";
import * as pulumi from "@pulumi/pulumi";

function getSnsPolicy(
  ingestionArn: string,
  snsArn: string
): aws.iam.PolicyDocument {
  return {
    Version: "2012-10-17",
    Statement: [
      {
        Effect: "Allow",

        Principal: {
          Service: "s3.amazonaws.com",
        },

        Action: ["SNS:Publish"],
        Resource: [snsArn],
        Condition: {
          ArnEquals: {
            "aws:SourceArn": ingestionArn,
          },
        },
      },
    ],
  };
}

function getSqsPolicy(snsArn: string, sqsArn: string): aws.iam.PolicyDocument {
  return {
    Version: "2012-10-17",
    Statement: [
      {
        Effect: "Allow",

        Principal: {
          Service: "s3.amazonaws.com",
        },

        Action: ["SQS:SendMessage"],
        Resource: [sqsArn],
        Condition: {
          ArnEquals: {
            "aws:SourceArn": snsArn,
          },
        },
      },
    ],
  };
}

const ingestion = new aws.s3.Bucket("prism-ingestion", {
  lifecycleRules: [
    {
      id: "ttl",
      prefix: "",
      enabled: true,
      expiration: {
        days: 3,
      },
    },
  ],
});

const sns = new aws.sns.Topic("prism-ingestion-sns");
const topicPolicy = new aws.sns.TopicPolicy("prism-ingestion-sns-policy", {
  arn: sns.arn,
  policy: pulumi
    .all([ingestion.arn, sns.arn])
    .apply(([ingestionArn, snsArn]) => getSnsPolicy(ingestionArn, snsArn))
    .apply(JSON.stringify),
});

const sqs = new aws.sqs.Queue("prism-ingestion-sqs");
const sqsPolicyAttachment = new aws.sqs.QueuePolicy(
  "prism-ingestion-sqs-policy",
  {
    queueUrl: sqs.url,
    policy: pulumi
      .all([sns.arn, sqs.arn])
      .apply(([snsArn, sqsArn]) => getSqsPolicy(snsArn, sqsArn))
      .apply(JSON.stringify),
  }
);

const subscription = new aws.sns.TopicSubscription(
  "prism-ingestion-sns-subscription",
  {
    topic: sns,
    protocol: "sqs",
    endpoint: sqs.arn,
  }
);

const bucketNotification = new aws.s3.BucketNotification(
  "prism-ingestion-s3-notification",
  {
    bucket: ingestion.id,
    topics: [
      {
        topicArn: sns.arn,
        events: ["s3:ObjectCreated:*"],
      },
    ],
  }
);

const storage = new aws.s3.Bucket("prism-storage");

export const ingestionBucket = ingestion.id;
export const storageBucket = storage.id;
export const sqsUrl = sqs.url;
