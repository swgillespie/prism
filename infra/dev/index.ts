import * as aws from "@pulumi/aws";

import { Ingestion } from "./ingestion";

const ingestion = new Ingestion("ingestion");
const storage = new aws.s3.Bucket("prism-storage");

export const ingestionBucket = ingestion.ingestionBucket;
export const storageBucket = storage.id;
export const sqsUrl = ingestion.sqsUrl;
