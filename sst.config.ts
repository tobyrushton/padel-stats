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

    let db;
    if ($app.stage === "production") {
      const bexboxPL = new neon.Project("bexbox-pl", {
        name: "bexbox-pl",
        regionId: "aws-eu-west-2",
        pgVersion: 17,
        orgId: process.env.NEON_ORG_ID,
        historyRetentionSeconds: 21600,
      })

      db = new sst.Linkable("NeonDB", {
        properties: {
          connectionString: bexboxPL.connectionUri,
        }
      });
    } else {
      db = new sst.Linkable("NeonDB", {
        properties: {
          connectionString: "postgresql://admin:admin@localhost:5432/mydb?sslmode=disable",
        }
      })
    }

    const api = new sst.aws.ApiGatewayV2("API", {
      link: [db],
    })

    api.route("$default", {
      handler: "github.com/tobyrushton/padel-stats/pkg/api/",
      runtime: "go",
      environment: {
        JWT_SECRET: process.env.JWT_SECRET!,
        JWT_ISSUER: process.env.JWT_ISSUER!,
      }
    })

    const web = new sst.aws.Astro("www", {
      link: [api],
      path: "pkg/www",
    })
  },
});
