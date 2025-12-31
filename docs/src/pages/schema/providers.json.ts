export function GET() {
  return new Response(
    JSON.stringify({
      $schema: "https://json-schema.org/draft/2020-12/schema",
      $id: "providers.schema.json",
      title: "Providers",
      description: "Configure which provider instances are enabled in the UI.",
      type: "object",
      properties: {
        include: {
          title: "Include Providers",
          description:
            "List of provider patterns to include. Supports exact instance IDs (gitlab:gitlab.com), provider wildcards (gitlab:*), or provider aliases (gitlab/github).",
          type: "array",
          items: {
            type: "string",
          },
        },
        exclude: {
          title: "Exclude Providers",
          description:
            "List of provider patterns to exclude. Supports exact instance IDs (gitlab:gitlab.com), provider wildcards (gitlab:*), or provider aliases (gitlab/github).",
          type: "array",
          items: {
            type: "string",
          },
        },
        defaults: {
          title: "Provider Defaults",
          description:
            "Default UI behavior for provider-aware views.",
          type: "object",
          properties: {
            groupByProvider: {
              title: "Group By Provider",
              description:
                "When true, sections are grouped by provider instance by default.",
              type: "boolean",
            },
          },
        },
      },
    }),
  );
}
