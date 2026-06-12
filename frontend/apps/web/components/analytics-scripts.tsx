'use client'

// Deep import (not the package barrel) to keep axios out of this client chunk.
import { extractTenantFromHostname } from '@workspace/shared-http/tenant-utils'
import Script from 'next/script'
import { useEffect, useState } from 'react'

import { env } from '@/lib/env'

/**
 * GA + Meta Pixel, only on the apex (public marketing site) — never on tenant
 * subdomains (the authenticated app surface).
 *
 * Client component on purpose: the apex check needs the request hostname, and
 * reading it server-side (next/headers) is a dynamic API that breaks static/ISR
 * rendering of every page under the root layout (digest DYNAMIC_SERVER_USAGE).
 * In the browser the hostname is free, and the scripts are afterInteractive
 * anyway, so nothing is lost by deciding after mount.
 */
export function AnalyticsScripts() {
  const [isApex, setIsApex] = useState(false)

  useEffect(() => {
    setIsApex(extractTenantFromHostname(window.location.hostname) === null)
  }, [])

  const gaId = env.NEXT_PUBLIC_GA_ID
  const metaPixelId = env.NEXT_PUBLIC_META_PIXEL_ID

  if (!isApex) return null

  return (
    <>
      {gaId && (
        <>
          <Script
            src={`https://www.googletagmanager.com/gtag/js?id=${gaId}`}
            strategy="afterInteractive"
          />
          {/* biome-ignore lint/correctness/useUniqueElementIds: next/script inline tag needs a stable, single-use id for GA */}
          <Script id="ga-gtag" strategy="afterInteractive">
            {`window.dataLayer = window.dataLayer || [];
function gtag(){dataLayer.push(arguments);}
gtag('js', new Date());
gtag('config', '${gaId}');`}
          </Script>
        </>
      )}
      {metaPixelId && (
        <>
          {/* biome-ignore lint/correctness/useUniqueElementIds: next/script inline tag needs a stable, single-use id for the Meta Pixel */}
          <Script id="meta-pixel" strategy="afterInteractive">
            {`!function(f,b,e,v,n,t,s)
{if(f.fbq)return;n=f.fbq=function(){n.callMethod?
n.callMethod.apply(n,arguments):n.queue.push(arguments)};
if(!f._fbq)f._fbq=n;n.push=n;n.loaded=!0;n.version='2.0';
n.queue=[];t=b.createElement(e);t.async=!0;
t.src=v;s=b.getElementsByTagName(e)[0];
s.parentNode.insertBefore(t,s)}(window, document,'script',
'https://connect.facebook.net/en_US/fbevents.js');
fbq('init', '${metaPixelId}');
fbq('track', 'PageView');`}
          </Script>
        </>
      )}
    </>
  )
}
