import { PageDetailFeature } from '@/features/page-detail'

interface PageProps {
  params: Promise<{
    id: string
    pageId: string
  }>
}

export default async function PageDetailPage(props: Readonly<PageProps>) {
  const params = await props.params
  const { id: workspaceId, pageId } = params

  return <PageDetailFeature workspaceId={workspaceId} pageId={pageId} />
}
