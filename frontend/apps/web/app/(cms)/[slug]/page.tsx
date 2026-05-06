import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { BlocksRenderer } from '@/features/cms'
import { FooterSection, Navbar } from '@/features/landing'
import { getPayloadClient } from '@/lib/payload'

export const revalidate = 60

type Props = {
  params: Promise<{ slug: string }>
}

export async function generateStaticParams() {
  const payload = await getPayloadClient()
  const pages = await payload.find({
    collection: 'pages',
    where: { _status: { equals: 'published' } },
    select: { slug: true },
    limit: 1000,
  })
  return pages.docs.map((page: any) => ({ slug: page.slug }))
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { slug } = await params
  const payload = await getPayloadClient()
  const result = await payload.find({
    collection: 'pages',
    where: { slug: { equals: slug }, _status: { equals: 'published' } },
    limit: 1,
  })
  const page = result.docs[0] as any
  if (!page) return {}
  return {
    title: page.meta?.title ?? page.title,
    description: page.meta?.description ?? undefined,
  }
}

export default async function DynamicPage({ params }: Props) {
  const { slug } = await params
  const payload = await getPayloadClient()
  const result = await payload.find({
    collection: 'pages',
    where: { slug: { equals: slug }, _status: { equals: 'published' } },
    limit: 1,
    depth: 2,
  })
  const page = result.docs[0] as any
  if (!page) notFound()

  return (
    <div className="min-h-screen bg-[#f3f3f3]">
      <div className="mx-auto max-w-[1280px] space-y-3 p-3">
        <Navbar />
        <main>
          <BlocksRenderer blocks={page.blocks ?? []} />
        </main>
        <FooterSection />
      </div>
    </div>
  )
}
