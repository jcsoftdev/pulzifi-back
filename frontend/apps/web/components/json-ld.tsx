/**
 * JsonLd — RSC that injects a JSON-LD structured-data script tag.
 * Accepts any valid schema.org object; serialisation happens server-side.
 */
export function JsonLd({ data }: { data: Record<string, unknown> }) {
  return (
    <script
      type="application/ld+json"
      // biome-ignore lint/security/noDangerouslySetInnerHtml: intentional structured-data injection — server-side only, no user input
      dangerouslySetInnerHTML={{ __html: JSON.stringify(data) }}
    />
  )
}
