import * as aws from "@pulumi/aws";
import * as k8s from "@pulumi/kubernetes";

import { Ingestion } from "./ingestion";
import { Cluster } from "./cluster";
import { Base } from "./k8s/base";
import { MetaService } from "./k8s/meta";

const ingestion = new Ingestion("ingestion");
const storage = new aws.s3.Bucket("prism-storage", { forceDestroy: true });
const cluster = new Cluster("prism");

const k8sProvider = new k8s.Provider("cluster-provider", {
  kubeconfig: cluster.kubeconfig,
});

const base = new Base("base", {
  provider: k8sProvider,
});

const meta = new MetaService("meta", {
  provider: k8sProvider,
  namespace: base.namespace,
});

export const ingestionBucket = ingestion.ingestionBucket;
export const storageBucket = storage.id;
export const sqsUrl = ingestion.sqsUrl;
export const kubeconfig = cluster.kubeconfig;
