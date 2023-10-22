import * as pulumi from "@pulumi/pulumi";
import * as aws from "@pulumi/aws";

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

export class Ingestion extends pulumi.ComponentResource {
  public readonly ingestionBucket: pulumi.Output<string>;
  public readonly sqsUrl: pulumi.Output<string>;

  constructor(name: string) {
    super("prism:ingestion:Ingestion", name);

    const ingestBucket = new aws.s3.Bucket(
      "prism-ingestion",
      {
        forceDestroy: true,
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
      },
      { parent: this }
    );

    const sns = new aws.sns.Topic("prism-ingestion-sns", {}, { parent: this });
    const topicPolicy = new aws.sns.TopicPolicy(
      "prism-ingestion-sns-policy",
      {
        arn: sns.arn,
        policy: pulumi
          .all([ingestBucket.arn, sns.arn])
          .apply(([ingestionArn, snsArn]) => getSnsPolicy(ingestionArn, snsArn))
          .apply(JSON.stringify),
      },
      { parent: this }
    );

    const sqs = new aws.sqs.Queue("prism-ingestion-sqs");
    const sqsPolicyAttachment = new aws.sqs.QueuePolicy(
      "prism-ingestion-sqs-policy",
      {
        queueUrl: sqs.url,
        policy: pulumi
          .all([sns.arn, sqs.arn])
          .apply(([snsArn, sqsArn]) => getSqsPolicy(snsArn, sqsArn))
          .apply(JSON.stringify),
      },
      { parent: this }
    );

    const subscription = new aws.sns.TopicSubscription(
      "prism-ingestion-sns-subscription",
      {
        topic: sns,
        protocol: "sqs",
        endpoint: sqs.arn,
      },
      { parent: this }
    );

    const bucketNotification = new aws.s3.BucketNotification(
      "prism-ingestion-s3-notification",
      {
        bucket: ingestBucket.id,
        topics: [
          {
            topicArn: sns.arn,
            events: ["s3:ObjectCreated:*"],
          },
        ],
      },
      { parent: this, dependsOn: [sqs] }
    );

    this.ingestionBucket = ingestBucket.id;
    this.sqsUrl = sqs.url;
    this.registerOutputs({
      ingestionBucket: this.ingestionBucket,
      sqsUrl: this.sqsUrl,
    });
  }
}
