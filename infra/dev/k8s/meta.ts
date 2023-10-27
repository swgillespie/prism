import * as pulumi from "@pulumi/pulumi";
import * as k8s from "@pulumi/kubernetes";
import * as awsx from "@pulumi/awsx";

export interface MetaServiceArgs {
  provider: k8s.Provider;
  namespace: k8s.core.v1.Namespace;
}

export class MetaService extends pulumi.ComponentResource {
  constructor(
    name: string,
    args: MetaServiceArgs,
    opts?: pulumi.ComponentResourceOptions
  ) {
    super("prism:meta:MetaService", name, {}, opts);
    const childOpts: pulumi.ComponentResourceOptions = {
      parent: this,
      provider: args.provider,
      ...opts,
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
        dockerfile: "../../go/services/prism-meta/Dockerfile",
      },
      childOpts
    );

    const config = new pulumi.Config("cockroachdb");
    const secret = new k8s.core.v1.Secret(
      `${name}-secret`,
      {
        metadata: {
          namespace: args.namespace.metadata.name,
          name: "cockroachdb",
        },
        stringData: {
          password: config.requireSecret("password"),
        },
      },
      childOpts
    );

    const labels = { app: "prism-meta" };
    const deployment = new k8s.apps.v1.Deployment(
      `${name}-deployment`,
      {
        metadata: {
          namespace: args.namespace.metadata.name,
          name: "meta",
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
                  name: "prism-meta",
                  env: [
                    {
                      name: "COCKROACHDB_USER",
                      value: config.require("user"),
                    },
                    {
                      name: "COCKROACHDB_PASSWORD",
                      valueFrom: {
                        secretKeyRef: {
                          name: secret.metadata.name,
                          key: "password",
                        },
                      },
                    },
                    {
                      name: "COCKROACHDB_URL",
                      value: config.require("url"),
                    },
                    {
                      name: "COCKROACHDB_DATABASE",
                      value: config.require("database"),
                    },
                  ],
                  resources: {
                    requests: {
                      cpu: "100m",
                      memory: "128Mi",
                    },
                    limits: {
                      cpu: "500m",
                      memory: "256Mi",
                    },
                  },
                  ports: [
                    {
                      name: "http",
                      containerPort: 8080,
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
          namespace: args.namespace.metadata.name,
          name: "meta",
          labels: labels,
        },
        spec: {
          selector: labels,
          clusterIP: "None",
          ports: [
            {
              name: "http",
              port: 8080,
              targetPort: 8080,
            },
          ],
        },
      },
      childOpts
    );

    this.registerOutputs({
      service,
    });
  }
}
