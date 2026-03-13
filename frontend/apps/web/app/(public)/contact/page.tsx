'use client'

import { Button } from '@workspace/ui/components/atoms/button'
import { cn } from '@workspace/ui/lib/utils'
import { CheckCircle, Linkedin, Twitter, Youtube } from 'lucide-react'
import { useId, useState } from 'react'
import { FooterSection } from '@/features/landing/ui/footer-section'
import { Navbar } from '@/features/landing/ui/navbar'

const inputClass =
  'h-14 w-full rounded-full bg-[#f3f3f3] px-6 text-base font-medium leading-6 tracking-[-0.128px] text-[#111] outline-none placeholder:text-[#111]/40 focus:ring-2 focus:ring-primary/30'

export default function ContactPage() {
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [message, setMessage] = useState('')
  const [submitted, setSubmitted] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const nameId = useId()
  const emailId = useId()
  const messageId = useId()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsLoading(true)

    // mailto fallback until a backend endpoint is available
    const subject = encodeURIComponent(`Custom Plan Inquiry from ${name}`)
    const body = encodeURIComponent(`Name: ${name}\nEmail: ${email}\n\n${message}`)
    window.open(`mailto:support@pulzifi.com?subject=${subject}&body=${body}`, '_self')

    setSubmitted(true)
    setIsLoading(false)
  }

  return (
    <div className="min-h-screen bg-[#f3f3f3]">
      <div className="mx-auto max-w-[1280px] space-y-3 p-3">
        <Navbar />

        <section className="mx-auto max-w-[1256px] overflow-hidden rounded-3xl bg-white px-6 py-12 md:px-[58px] md:py-[50px]">
          <div className="flex flex-col items-center gap-[60px]">
            {/* Header */}
            <div className="flex max-w-[852px] flex-col items-center gap-4 text-center">
              <h1 className="font-heading text-5xl font-medium leading-[72px] tracking-[-3.6px] text-[#111] md:text-[60px]">
                Contact <em className="font-heading italic">Us</em>
              </h1>
              <p className="max-w-[494px] text-base leading-6 text-[#131313]">
                We&apos;re here to help! Whether you have questions, feedback, or need support, our
                team is ready to assist you.
              </p>
            </div>

            {/* Two-column layout */}
            <div className="flex w-full flex-col gap-16 lg:flex-row lg:gap-20">
              {/* Left: Get in touch */}
              <div className="flex flex-1 flex-col gap-6">
                <h2 className="font-heading text-5xl font-medium leading-[72px] tracking-[-3.6px] text-[#131313] md:text-[60px]">
                  Get in touch
                </h2>

                <div className="flex flex-col gap-6">
                  {/* Email */}
                  <div className="flex flex-col gap-3.5">
                    <span className="text-xl leading-7 text-[#111]/50">Email:</span>
                    <a
                      href="mailto:support@pulzifi.com"
                      className="text-2xl font-medium leading-8 text-[#111] hover:underline"
                    >
                      support@pulzifi.com
                    </a>
                  </div>

                  {/* Address */}
                  <div className="flex flex-col gap-3.5">
                    <span className="text-xl leading-7 text-[#111]/50">Address:</span>
                    <span className="text-2xl font-medium leading-8 text-[#111]">Boise, ID</span>
                  </div>

                  {/* Follow Us */}
                  <div className="flex flex-col gap-3.5">
                    <span className="text-xl leading-7 text-[#111]/50">Follow Us</span>
                    <div className="flex gap-6">
                      <a
                        href="https://x.com/pulzifi"
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-[#111] transition-colors hover:text-[#111]/70"
                        aria-label="X (Twitter)"
                      >
                        <Twitter className="size-6" />
                      </a>
                      <a
                        href="https://linkedin.com/company/pulzifi"
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-[#111] transition-colors hover:text-[#111]/70"
                        aria-label="LinkedIn"
                      >
                        <Linkedin className="size-6" />
                      </a>
                      <a
                        href="https://youtube.com/@pulzifi"
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-[#111] transition-colors hover:text-[#111]/70"
                        aria-label="YouTube"
                      >
                        <Youtube className="size-6" />
                      </a>
                    </div>
                  </div>
                </div>
              </div>

              {/* Right: Contact form */}
              <div className="flex flex-1 flex-col gap-[30px]">
                {submitted ? (
                  <div className="flex flex-1 flex-col items-center justify-center gap-4 text-center">
                    <CheckCircle className="size-12 text-green-500" />
                    <h3 className="text-2xl font-medium text-[#111]">Message sent!</h3>
                    <p className="text-base text-[#111]/50">
                      Thank you for reaching out. We&apos;ll get back to you shortly.
                    </p>
                    <button
                      type="button"
                      onClick={() => {
                        setSubmitted(false)
                        setName('')
                        setEmail('')
                        setMessage('')
                      }}
                      className="text-sm font-medium text-[#29144c] hover:underline"
                    >
                      Send another message
                    </button>
                  </div>
                ) : (
                  <form onSubmit={handleSubmit} className="flex flex-1 flex-col gap-[30px]">
                    {/* Name + Email row */}
                    <div className="flex flex-col gap-3.5 sm:flex-row">
                      <div className="flex flex-1 flex-col gap-3">
                        <label
                          htmlFor={nameId}
                          className="text-base font-medium leading-6 tracking-[-0.128px] text-[#111]"
                        >
                          Your Name
                        </label>
                        <input
                          id={nameId}
                          type="text"
                          value={name}
                          onChange={(e) => setName(e.target.value)}
                          required
                          placeholder="Your name"
                          className={inputClass}
                        />
                      </div>
                      <div className="flex flex-1 flex-col gap-3">
                        <label
                          htmlFor={emailId}
                          className="text-base font-medium leading-6 tracking-[-0.128px] text-[#111]"
                        >
                          Email address
                        </label>
                        <input
                          id={emailId}
                          type="email"
                          value={email}
                          onChange={(e) => setEmail(e.target.value)}
                          required
                          placeholder="Your email address"
                          className={inputClass}
                        />
                      </div>
                    </div>

                    {/* Message */}
                    <div className="flex flex-1 flex-col gap-3">
                      <label
                        htmlFor={messageId}
                        className="text-base font-medium leading-6 tracking-[-0.128px] text-[#111]"
                      >
                        Message
                      </label>
                      <textarea
                        id={messageId}
                        value={message}
                        onChange={(e) => setMessage(e.target.value)}
                        required
                        placeholder="Write something...."
                        className={cn(
                          'min-h-[200px] flex-1 resize-none rounded-3xl bg-[#f3f3f3] px-6 py-4 text-base font-medium leading-6 tracking-[-0.128px] text-[#111] outline-none placeholder:text-[#111]/40 focus:ring-2 focus:ring-primary/30'
                        )}
                      />
                    </div>

                    {/* Submit */}
                    <Button
                      type="submit"
                      disabled={isLoading}
                      className="h-14 w-full rounded-full bg-[#29144c] text-base font-medium tracking-[-0.128px] hover:bg-[#3d1d6e]"
                    >
                      {isLoading ? 'Sending...' : 'Send Message'}
                    </Button>
                  </form>
                )}
              </div>
            </div>
          </div>
        </section>

        <FooterSection />
      </div>
    </div>
  )
}
