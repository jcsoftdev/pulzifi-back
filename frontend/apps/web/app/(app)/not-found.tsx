import { ArrowRight } from 'lucide-react'
import Image from 'next/image'
import Link from 'next/link'

import { FooterSection, Navbar } from '@/features/landing'

export default function NotFound() {
  return (
    <div className="min-h-screen bg-[#ededed]">
      <div className="mx-auto max-w-[1280px] space-y-3 p-3">
        <Navbar />
        <main>
          <section className="relative flex min-h-[500px] items-center justify-center overflow-hidden rounded-3xl bg-white md:min-h-[700px]">
            <Image
              src="/images/landing/404-bg.svg"
              alt=""
              width={854}
              height={323}
              className="pointer-events-none absolute select-none"
              priority
              aria-hidden="true"
            />
            <div className="relative z-10 flex flex-col items-center gap-6 px-6">
              <div className="flex flex-col items-center gap-2">
                <h1 className="font-heading text-5xl font-medium leading-[56px] tracking-[-2.88px] text-[#131313] max-md:text-3xl max-md:leading-10">
                  Page not found
                </h1>
                <p className="text-base leading-6 text-[#444141]">
                  Let&apos;s try a different way...
                </p>
              </div>
              <Link
                href="/"
                className="inline-flex h-14 items-center gap-1 rounded-full bg-[#29144c] px-6 text-base font-medium capitalize leading-6 tracking-[-0.128px] text-white transition-all duration-300 hover:bg-[#3d1d6e] hover:scale-[1.02] active:scale-[0.98]"
              >
                Go to Home
                <ArrowRight className="size-6" />
              </Link>
            </div>
          </section>
        </main>
        <FooterSection />
      </div>
    </div>
  )
}
