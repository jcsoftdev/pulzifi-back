import { AuthApi } from '@workspace/services'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@workspace/ui/components/atoms/card'
import Link from 'next/link'
import { AcceptPlatformInviteForm } from '@/features/auth/ui/accept-platform-invite-form'

export const dynamic = 'force-dynamic'

interface PageProps {
  params: Promise<{
    token: string
  }>
}

export default async function AcceptPlatformInvitePage({ params }: PageProps) {
  const { token } = await params

  if (!token) {
    return <InvalidInvitation />
  }

  try {
    const invitation = await AuthApi.getInvitation(token)
    return (
      <div className="min-h-screen bg-[#f3f3f3]">
        <div className="mx-auto max-w-[1280px] space-y-3 p-3">
          <section className="relative mx-auto max-w-[1256px] overflow-hidden rounded-3xl bg-white px-6 py-12 md:px-0 md:py-[50px]">
            <div className="mx-auto flex w-full max-w-[583px] flex-col gap-[40px]">
              <h1 className="font-heading text-[48px] font-medium italic leading-[56px] tracking-[-1px] text-[#111] max-md:text-4xl max-md:leading-[44px]">
                You&apos;re invited.
              </h1>
              <AcceptPlatformInviteForm
                token={token}
                email={invitation.email}
                invitedByName={invitation.invitedByName}
              />
            </div>
          </section>
        </div>
      </div>
    )
  } catch {
    return <InvalidInvitation />
  }
}

function InvalidInvitation() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-background px-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl">Invitation unavailable</CardTitle>
          <CardDescription>This invitation is invalid or has expired.</CardDescription>
        </CardHeader>
        <CardContent className="text-center">
          <Link href="/login" className="text-primary hover:underline font-medium">
            Go to login
          </Link>
        </CardContent>
      </Card>
    </div>
  )
}
