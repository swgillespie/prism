import * as awsx from "@pulumi/awsx";
import * as pulumi from "@pulumi/pulumi";
import * as eks from "@pulumi/eks";

export class Cluster extends pulumi.ComponentResource {
  public readonly kubeconfig: pulumi.Output<any>;

  constructor(name: string, opts?: pulumi.ComponentResourceOptions) {
    super("prism:cluster:Cluster", name, opts);

    const vpc = new awsx.ec2.Vpc(`${name}-vpc`, {}, { parent: this, ...opts });
    const cluster = new eks.Cluster(
      `${name}-cluster`,
      {
        desiredCapacity: 3,
        minSize: 3,
        maxSize: 5,
        instanceType: "t2.medium",
        vpcId: vpc.vpcId,
        publicSubnetIds: vpc.publicSubnetIds,
        privateSubnetIds: vpc.privateSubnetIds,
        nodeAssociatePublicIpAddress: false,
      },
      { parent: this, ...opts }
    );

    this.kubeconfig = cluster.kubeconfig;
    this.registerOutputs({
      kubeconfig: this.kubeconfig,
    });
  }
}
