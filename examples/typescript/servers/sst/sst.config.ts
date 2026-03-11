/// <reference path=".sst/platform/config.d.ts" />

export default $config({
  app(input) {
    return {
      name: "x402-sst",
      removal: input?.stage === "production" ? "retain" : "remove",
      protect: ["production"],
      home: "aws",
    };
  },
  async run() {
    const api = new sst.aws.ApiGatewayV2("x402Api");

    api.route("$default", {
      handler: "src/index.handler",
      environment: {
        EVM_ADDRESS: process.env.EVM_ADDRESS ?? "",
        SVM_ADDRESS: process.env.SVM_ADDRESS ?? "",
        AVM_ADDRESS: process.env.AVM_ADDRESS ?? "",
        FACILITATOR_URL:
          process.env.FACILITATOR_URL ?? "https://facilitator.goplausible.xyz",
      },
    });

    return {
      url: api.url,
    };
  },
});
