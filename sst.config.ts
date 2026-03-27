/// <reference path="./.sst/platform/config.d.ts" />
export default $config({
  app(input) {
    return {
      name: "bexbox-pl",
      removal: input?.stage === "production" ? "retain" : "remove",
      protect: ["production"].includes(input?.stage),
      home: "aws",
      providers: { neon: "0.13.0" },
    };
  },
  async run() {
    const bexboxPL = new neon.Project("bexbox-pl", {
      name: "bexbox-pl",
      regionId: "aws-eu-west-2",
      pgVersion: 17,
      orgId: process.env.NEON_ORG_ID,
      historyRetentionSeconds: 21600,
    })

    const db = new sst.Linkable("NeonDB", {
      properties: {
        connectionString: bexboxPL.connectionUri,
      }
    });

    const api = new sst.aws.ApiGatewayV2("API", {
      link: [db],
    })

    api.route("$default", {
      handler: "pkg/api/",
      runtime: "go",
    })
  },
});
