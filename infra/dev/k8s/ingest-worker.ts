import * as pulumi from "@pulumi/pulumi";
import * as k8s from "@pulumi/kubernetes";
import * as awsx from "@pulumi/awsx";

export interface IngestWorkerServiceArgs {
  provider: k8s.Provider;
  namespace: k8s.core.v1.Namespace;
}

export class IngestWorkerService extends pulumi.ComponentResource {
  constructor(
    name: string,
    args: IngestWorkerServiceArgs,
    opts?: pulumi.ComponentResourceOptions
  ) {
    super("prism:meta:IngestWorkerService", name, args, opts);
    const childOpts: pulumi.ComponentResourceOptions = {
      ...opts,
      parent: this,
      provider: args.provider,
    };

    const ecr = new awsx.ecr.Repository(
      `${name}-repo`,
      {
        forceDelete: true,
      },
      childOpts
    );
    const image = new awsx.ecr.Image(
      `${name}-image`,
      {
        repositoryUrl: ecr.repository.repositoryUrl,
        path: "../..",
        dockerfile: "../../go/services/prism-ingest-worker/Dockerfile",
      },
      childOpts
    );

    const configMap = new k8s.core.v1.ConfigMap(
      `${name}-configmap`,
      {
        metadata: {
          namespace: args.namespace.metadata.name,
        },
        data: {
          "config.yaml": JSON.stringify({
            meta: {
              endpoint: "meta.prism.svc.cluster.local.:8080",
            },
            temporal: {
              endpoint: "temporal.temporal.svc.cluster.local.:7233",
              task_queue: "prism-ingest-worker",
            },
          }),
        },
      },
      childOpts
    );

    const labels = { app: "prism-ingest-worker" };
    const deployment = new k8s.apps.v1.Deployment(
      `${name}-deployment`,
      {
        metadata: {
          namespace: args.namespace.metadata.name,
          name: "ingest-worker",
          labels: labels,
        },
        spec: {
          replicas: 2,
          selector: {
            matchLabels: labels,
          },
          template: {
            metadata: {
              labels: labels,
            },
            spec: {
              containers: [
                {
                  image: image.imageUri,
                  name: "prism-ingest-worker",
                  resources: {
                    requests: {
                      cpu: "100m",
                      memory: "256Mi",
                    },
                    limits: {
                      cpu: "100m",
                      memory: "512Mi",
                    },
                  },
                  ports: [
                    {
                      name: "metrics",
                      containerPort: 9090,
                    },
                  ],
                  volumeMounts: [
                    {
                      name: "config",
                      mountPath: "/etc/prism",
                    },
                  ],
                },
              ],
              volumes: [
                {
                  name: "config",
                  configMap: {
                    name: configMap.metadata.name,
                  },
                },
              ],
            },
          },
        },
      },
      childOpts
    );

    const service = new k8s.core.v1.Service(
      `${name}-service`,
      {
        metadata: {
          namespace: args.namespace.metadata.name,
          name: "ingest-worker",
          labels: labels,
        },
        spec: {
          selector: labels,
          type: "ClusterIP",
          ports: [
            {
              name: "metrics",
              port: 9090,
              targetPort: 9090,
            },
          ],
        },
      },
      childOpts
    );

    this.registerOutputs({ deployment, service });
  }
}
