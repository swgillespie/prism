import * as pulumi from "@pulumi/pulumi";
import * as k8s from "@pulumi/kubernetes";

export interface BaseArgs {
  provider: k8s.Provider;
}

export class Base extends pulumi.ComponentResource {
  public readonly namespace: k8s.core.v1.Namespace;
  public readonly temporal: TemporalService;

  constructor(
    name: string,
    args: BaseArgs,
    opts?: pulumi.ComponentResourceOptions
  ) {
    super("prism:infra:Base", name, args, opts);
    const childOpts = { parent: this, provider: args.provider, ...opts };

    this.namespace = new k8s.core.v1.Namespace(
      `${name}-namespace`,
      {
        metadata: {
          name: "prism",
        },
      },
      childOpts
    );

    this.temporal = new TemporalService(
      `${name}-temporal`,
      {
        provider: args.provider,
      },
      childOpts
    );
    this.registerOutputs({
      namespace: this.namespace,
    });
  }
}

export interface TemporalServiceArgs {
  provider: k8s.Provider;
}

export class TemporalService extends pulumi.ComponentResource {
  constructor(
    name: string,
    args: TemporalServiceArgs,
    opts?: pulumi.ComponentResourceOptions
  ) {
    super("prism:infra:TemporalService", name, args, opts);
    const childOpts = { ...opts, parent: this, provider: args.provider };
    const namespace = new k8s.core.v1.Namespace(
      `${name}-namespace`,
      {
        metadata: {
          name: "temporal",
        },
      },
      childOpts
    );
    const labels = { app: "temporal" };
    const deployment = new k8s.apps.v1.Deployment(
      `${name}-deployment`,
      {
        metadata: {
          namespace: namespace.metadata.name,
          name: "temporal",
          labels: labels,
        },
        spec: {
          replicas: 1,
          selector: { matchLabels: labels },
          template: {
            metadata: { labels: labels },
            spec: {
              containers: [
                {
                  image: "temporalio/admin-tools:1.22.1.0",
                  name: "temporal",
                  command: ["temporal", "server", "start-dev"],
                  resources: {
                    requests: {
                      cpu: "100m",
                      memory: "128Mi",
                    },
                    limits: {
                      cpu: "100m",
                      memory: "128Mi",
                    },
                  },
                  ports: [
                    {
                      name: "grpc",
                      containerPort: 7233,
                    },
                    {
                      name: "http",
                      containerPort: 8233,
                    },
                  ],
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
          namespace: namespace.metadata.name,
          name: "temporal",
          labels: labels,
        },
        spec: {
          selector: labels,
          type: "ClusterIP",
          ports: [
            {
              name: "grpc",
              port: 7233,
              targetPort: 7233,
            },
            {
              name: "http",
              port: 8233,
              targetPort: 8233,
            },
          ],
        },
      },
      childOpts
    );
    this.registerOutputs({ deployment, service });
  }
}
