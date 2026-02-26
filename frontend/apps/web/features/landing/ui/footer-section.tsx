import Link from 'next/link'
import { FOOTER_LINKS } from '../lib/data'
import { LandingButton } from './components/landing-button'

export function FooterSection() {
  return (
    <footer
      id="footer"
      className="relative mx-auto max-w-[1256px] overflow-hidden rounded-3xl bg-[#050209] px-6 py-12 md:px-[58px] md:py-[50px]"
    >
      <div className="relative z-10 flex flex-col items-center gap-[50px]">
        {/* Newsletter */}
        <div className="flex flex-col items-center gap-[60px]">
          <div className="flex flex-col items-center gap-4 text-center">
            <h2 className="font-heading max-w-[520px] text-5xl font-medium leading-[56px] tracking-[-2.88px] text-white max-md:text-3xl max-md:leading-10">
              Learn more about how to take advantage with Pulzifi
            </h2>
            <p className="max-w-[504px] text-base leading-6 text-white/80">
              Subscribe to our newsletter for expert tips, updates, and the latest trends in
              monitoring and actions.
            </p>
          </div>
          <div className="flex w-full max-w-[500px] flex-col gap-2.5 sm:flex-row">
            <div className="flex flex-1 items-center rounded-full bg-white/10 px-6">
              <input
                type="email"
                placeholder="Enter your email"
                className="w-full bg-transparent py-4 text-base font-medium text-white outline-none placeholder:text-white/60"
                aria-label="Email for newsletter"
              />
            </div>
            <LandingButton
              href="#"
              className="bg-[#40078b] hover:bg-[#5b1aab] shadow-none w-full justify-center sm:w-auto"
              variant="primary"
              size="lg"
            >
              Subscribe
            </LandingButton>
          </div>
        </div>

        {/* Divider */}
        <div className="h-px w-full bg-white/10" />

        {/* Footer links */}
        <div className="flex w-full flex-col items-start justify-between gap-10 md:flex-row md:items-center">
          <div>
            <span className="font-logo text-7xl font-extrabold tracking-[0.67px] text-white max-md:text-5xl">
              Pulzifi
            </span>
          </div>
          <div className="flex flex-wrap gap-8 sm:gap-16">
            {Object.entries(FOOTER_LINKS).map(([category, links]) => (
              <div key={category} className="flex flex-col gap-4">
                <span className="text-base text-white">{category}</span>
                <div className="flex flex-col gap-3">
                  {links.map((link) => (
                    <Link
                      key={link.label}
                      href={link.href}
                      className="text-sm text-white/60 transition-colors hover:text-white"
                    >
                      {link.label}
                    </Link>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Large background Pulzifi text */}
      <span
        className="pointer-events-none absolute bottom-[-40px] left-4 select-none text-[clamp(80px,20vw,387px)] font-extrabold leading-none tracking-wide text-white/10 font-logo sm:bottom-[-80px] sm:left-12"
        aria-hidden="true"
      >
        Pulzifi
      </span>
    </footer>
  )
}
