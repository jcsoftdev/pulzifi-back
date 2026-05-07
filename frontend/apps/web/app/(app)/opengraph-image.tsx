import { ImageResponse } from 'next/og'

export const runtime = 'edge'
export const alt = 'Pulzifi — AI-Powered Competitive Intelligence'
export const size = { width: 1200, height: 630 }
export const contentType = 'image/png'

export default function Image() {
  return new ImageResponse(
    (
      <div
        style={{
          background: '#050508',
          width: '100%',
          height: '100%',
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          fontFamily: 'system-ui, sans-serif',
          position: 'relative',
          overflow: 'hidden',
        }}
      >
        {/* Background glow */}
        <div
          style={{
            position: 'absolute',
            top: '-80px',
            left: '50%',
            transform: 'translateX(-50%)',
            width: '700px',
            height: '500px',
            background:
              'radial-gradient(ellipse at center, rgba(124,58,237,0.18) 0%, transparent 70%)',
            borderRadius: '50%',
          }}
        />
        <div
          style={{
            position: 'absolute',
            bottom: '-60px',
            right: '100px',
            width: '350px',
            height: '350px',
            background:
              'radial-gradient(ellipse at center, rgba(99,102,241,0.1) 0%, transparent 70%)',
            borderRadius: '50%',
          }}
        />

        {/* Logo row */}
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '12px',
            marginBottom: '32px',
          }}
        >
          {/* Icon */}
          <div
            style={{
              width: '52px',
              height: '52px',
              borderRadius: '14px',
              background: 'linear-gradient(135deg, #7c3aed, #4f46e5)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            <svg
              width="28"
              height="28"
              viewBox="0 0 24 24"
              fill="none"
              stroke="white"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <path d="M12 2a10 10 0 1 0 10 10" />
              <path d="M12 6a6 6 0 1 0 6 6" />
              <path d="M12 10a2 2 0 1 0 2 2" />
            </svg>
          </div>
          <span
            style={{
              fontSize: '28px',
              fontWeight: 700,
              color: 'white',
              letterSpacing: '-0.5px',
            }}
          >
            Pulzifi
          </span>
        </div>

        {/* Badge */}
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '8px',
            background: 'rgba(124,58,237,0.12)',
            border: '1px solid rgba(124,58,237,0.3)',
            borderRadius: '100px',
            padding: '6px 16px',
            marginBottom: '28px',
          }}
        >
          <div
            style={{
              width: '6px',
              height: '6px',
              borderRadius: '50%',
              background: '#a78bfa',
            }}
          />
          <span
            style={{
              fontSize: '14px',
              fontWeight: 500,
              color: '#c4b5fd',
              letterSpacing: '0.3px',
            }}
          >
            AI-Powered Competitive Intelligence
          </span>
        </div>

        {/* Headline */}
        <div
          style={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            gap: '4px',
            marginBottom: '24px',
          }}
        >
          <span
            style={{
              fontSize: '54px',
              fontWeight: 800,
              color: 'white',
              letterSpacing: '-2px',
              lineHeight: 1.05,
            }}
          >
            Track What Competitors Do.
          </span>
          <span
            style={{
              fontSize: '54px',
              fontWeight: 800,
              letterSpacing: '-2px',
              lineHeight: 1.05,
              background: 'linear-gradient(90deg, #a78bfa, #818cf8, #60a5fa)',
              backgroundClip: 'text',
              color: 'transparent',
            }}
          >
            Act Before They Do.
          </span>
        </div>

        {/* Subtitle */}
        <p
          style={{
            fontSize: '18px',
            color: 'rgba(255,255,255,0.45)',
            margin: '0',
            maxWidth: '600px',
            textAlign: 'center',
            lineHeight: 1.5,
            marginBottom: '44px',
          }}
        >
          Monitor any website for changes and get AI-powered strategic insights
          delivered to your team — automatically, 24/7.
        </p>

        {/* Stats row */}
        <div style={{ display: 'flex', gap: '16px' }}>
          {[
            { label: 'Pages Monitored', value: '248+' },
            { label: 'Changes Detected', value: '12k+' },
            { label: 'AI Insights', value: '94+' },
          ].map(({ label, value }) => (
            <div
              key={label}
              style={{
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                background: 'rgba(255,255,255,0.04)',
                border: '1px solid rgba(255,255,255,0.08)',
                borderRadius: '12px',
                padding: '14px 28px',
              }}
            >
              <span
                style={{
                  fontSize: '24px',
                  fontWeight: 700,
                  color: 'white',
                  letterSpacing: '-0.5px',
                }}
              >
                {value}
              </span>
              <span
                style={{
                  fontSize: '12px',
                  color: 'rgba(255,255,255,0.35)',
                  marginTop: '2px',
                }}
              >
                {label}
              </span>
            </div>
          ))}
        </div>

        {/* Bottom URL */}
        <div
          style={{
            position: 'absolute',
            bottom: '28px',
            display: 'flex',
            alignItems: 'center',
            gap: '6px',
          }}
        >
          <span style={{ fontSize: '13px', color: 'rgba(255,255,255,0.2)' }}>
            pulzifi.com
          </span>
        </div>
      </div>
    ),
    { ...size },
  )
}
