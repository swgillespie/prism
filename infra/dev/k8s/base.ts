import * as pulumi from "@pulumi/pulumi";
import * as k8s from "@pulumi/kubernetes";

export interface BaseArgs {
  provider: k8s.Provider;
}

export class Base extends pulumi.ComponentResource {
  public readonly namespace: k8s.core.v1.Namespace;

  constructor(
    name: string,
    args: BaseArgs,
    opts?: pulumi.ComponentResourceOptions
  ) {
    super("prism:infra:Base", name, {}, opts);
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

    this.registerOutputs({
      namespace: this.namespace,
    });
  }
}
