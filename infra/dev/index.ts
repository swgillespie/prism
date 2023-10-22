import * as aws from "@pulumi/aws";

import { Ingestion } from "./ingestion";
import { Cluster } from "./cluster";

const ingestion = new Ingestion("ingestion");
const storage = new aws.s3.Bucket("prism-storage", { forceDestroy: true });
const cluster = new Cluster("prism");

export const ingestionBucket = ingestion.ingestionBucket;
export const storageBucket = storage.id;
export const sqsUrl = ingestion.sqsUrl;
export const kubeconfig = cluster.kubeconfig;
