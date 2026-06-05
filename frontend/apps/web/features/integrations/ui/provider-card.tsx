import type { Integration } from '../domain/types'
import { ConnectButton } from './connect-button'

interface ProviderCardProps {
  providerKey: string
  label: string
  description: string
  enabled: boolean
  authType: 'oauth' | 'byo' | 'none'
  soon?: boolean
  integration?: Integration
  planCode?: string
}

function ProviderIcon({ providerKey }: { providerKey: string }) {
  switch (providerKey) {
    case 'slack':
      return (
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="w-8 h-8"
          viewBox="0 0 24 24"
          fill="currentColor"
        >
          <title>Slack</title>
          <path d="M5.042 15.165a2.528 2.528 0 0 1-2.52 2.523A2.528 2.528 0 0 1 0 15.165a2.527 2.527 0 0 1 2.522-2.52h2.52v2.52zM6.313 15.165a2.527 2.527 0 0 1 2.521-2.52 2.527 2.527 0 0 1 2.521 2.52v6.313A2.528 2.528 0 0 1 8.834 24a2.528 2.528 0 0 1-2.521-2.522v-6.313zM8.834 5.042a2.528 2.528 0 0 1-2.521-2.52A2.528 2.528 0 0 1 8.834 0a2.528 2.528 0 0 1 2.521 2.522v2.52H8.834zM8.834 6.313a2.528 2.528 0 0 1 2.521 2.521 2.528 2.528 0 0 1-2.521 2.521H2.522A2.528 2.528 0 0 1 0 8.834a2.528 2.528 0 0 1 2.522-2.521h6.312zM18.956 8.834a2.528 2.528 0 0 1 2.522-2.521A2.528 2.528 0 0 1 24 8.834a2.528 2.528 0 0 1-2.522 2.521h-2.522V8.834zM17.688 8.834a2.528 2.528 0 0 1-2.523 2.521 2.527 2.527 0 0 1-2.52-2.521V2.522A2.527 2.527 0 0 1 15.165 0a2.528 2.528 0 0 1 2.523 2.522v6.312zM15.165 18.956a2.528 2.528 0 0 1 2.523 2.522A2.528 2.528 0 0 1 15.165 24a2.527 2.527 0 0 1-2.52-2.522v-2.522h2.52zM15.165 17.688a2.527 2.527 0 0 1-2.52-2.523 2.526 2.526 0 0 1 2.52-2.52h6.313A2.527 2.527 0 0 1 24 15.165a2.528 2.528 0 0 1-2.522 2.523h-6.313z" />
        </svg>
      )
    case 'email':
      return (
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="w-8 h-8"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          strokeWidth={1.5}
        >
          <title>Email</title>
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            d="M21.75 6.75v10.5a2.25 2.25 0 01-2.25 2.25h-15a2.25 2.25 0 01-2.25-2.25V6.75m19.5 0A2.25 2.25 0 0019.5 4.5h-15a2.25 2.25 0 00-2.25 2.25m19.5 0v.243a2.25 2.25 0 01-1.07 1.916l-7.5 4.615a2.25 2.25 0 01-2.36 0L3.32 8.91a2.25 2.25 0 01-1.07-1.916V6.75"
          />
        </svg>
      )
    case 'gmail':
      return (
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="w-8 h-8"
          viewBox="0 0 24 24"
          fill="currentColor"
        >
          <title>Gmail</title>
          <path d="M24 5.457v13.909c0 .904-.732 1.636-1.636 1.636h-3.819V11.73L12 16.64l-6.545-4.91v9.273H1.636A1.636 1.636 0 0 1 0 19.366V5.457c0-2.023 2.309-3.178 3.927-1.964L5.455 4.64 12 9.548l6.545-4.907 1.528-1.148C21.69 2.28 24 3.434 24 5.457z" />
        </svg>
      )
    case 'outlook':
      return (
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="w-8 h-8"
          viewBox="0 0 24 24"
          fill="currentColor"
        >
          <title>Outlook</title>
          <path d="M7.88 12.04q0 .45-.11.87-.1.41-.33.74-.22.33-.58.52-.37.2-.87.2t-.85-.2q-.35-.21-.57-.55-.22-.33-.33-.75-.1-.42-.1-.86t.1-.87q.1-.43.34-.76.22-.34.59-.54.36-.2.87-.2t.86.2q.35.21.57.55.22.34.31.77.1.43.1.88zM24 12v9.38q0 .46-.33.8-.33.32-.8.32H7.13q-.46 0-.8-.33-.32-.33-.32-.8V18H1q-.41 0-.7-.3-.3-.29-.3-.7V7q0-.41.3-.7Q.58 6 1 6h6.5V2.55q0-.44.3-.75.3-.3.75-.3h12.9q.44 0 .75.3.3.3.3.75V10.85l1.24.72h.01q.1.07.18.18.07.12.07.25zm-6-8.2v1.95l2 1.2V3.8zm-6.8 6.2q-.4-1.2-1.44-1.87-.47-.32-1.03-.48-.56-.17-1.15-.17-1.04 0-1.89.4-.56.27-.97.72-.4.45-.63 1.04-.23.59-.23 1.27 0 .66.23 1.25.22.59.63 1.04.4.45.97.72.85.4 1.89.4.97 0 1.73-.35.76-.36 1.25-.97.5-.62.74-1.44.1-.41.1-.87 0-.35-.06-.69-.06-.34-.14-.66z" />
        </svg>
      )
    case 'teams':
      return (
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="w-8 h-8"
          viewBox="0 0 24 24"
          fill="currentColor"
        >
          <title>Teams</title>
          <path d="M20.625 7.125a2.625 2.625 0 1 0 0-5.25 2.625 2.625 0 0 0 0 5.25zm-5.08 2.277a4.5 4.5 0 0 0-6.09 0 4.5 4.5 0 0 0 0 6.09 4.5 4.5 0 0 0 6.09 0 4.5 4.5 0 0 0 0-6.09zm-3.045 7.89a3.375 3.375 0 1 1 0-6.75 3.375 3.375 0 0 1 0 6.75zm7.687-6.792H18c-.207 0-.367.16-.367.367v4.633c0 1.013-.82 1.833-1.833 1.833s-1.833-.82-1.833-1.833V11.1a5.625 5.625 0 0 1-1.717-1.282 4.883 4.883 0 0 1 1.05-7.143A5.576 5.576 0 0 1 15.5 2h4.75C21.216 2 22 2.784 22 3.75v3.75c0 .414-.336.75-.75.75h-.063V10.5z" />
        </svg>
      )
    case 'discord':
      return (
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="w-8 h-8"
          viewBox="0 0 24 24"
          fill="currentColor"
        >
          <title>Discord</title>
          <path d="M20.317 4.37a19.791 19.791 0 0 0-4.885-1.515.074.074 0 0 0-.079.037c-.21.375-.444.864-.608 1.25a18.27 18.27 0 0 0-5.487 0 12.64 12.64 0 0 0-.617-1.25.077.077 0 0 0-.079-.037A19.736 19.736 0 0 0 3.677 4.37a.07.07 0 0 0-.032.027C.533 9.046-.32 13.58.099 18.057c.002.022.013.04.031.053a19.9 19.9 0 0 0 5.993 3.03.078.078 0 0 0 .084-.028 14.09 14.09 0 0 0 1.226-1.994.076.076 0 0 0-.041-.106 13.107 13.107 0 0 1-1.872-.892.077.077 0 0 1-.008-.128 10.2 10.2 0 0 0 .372-.292.074.074 0 0 1 .077-.01c3.928 1.793 8.18 1.793 12.062 0a.074.074 0 0 1 .078.01c.12.098.246.198.373.292a.077.077 0 0 1-.006.127 12.299 12.299 0 0 1-1.873.892.077.077 0 0 0-.041.107c.36.698.772 1.362 1.225 1.993a.076.076 0 0 0 .084.028 19.839 19.839 0 0 0 6.002-3.03.077.077 0 0 0 .032-.054c.5-5.177-.838-9.674-3.549-13.66a.061.061 0 0 0-.031-.03zM8.02 15.33c-1.183 0-2.157-1.085-2.157-2.419 0-1.333.956-2.419 2.157-2.419 1.21 0 2.176 1.096 2.157 2.42 0 1.333-.956 2.418-2.157 2.418zm7.975 0c-1.183 0-2.157-1.085-2.157-2.419 0-1.333.955-2.419 2.157-2.419 1.21 0 2.176 1.096 2.157 2.42 0 1.333-.946 2.418-2.157 2.418z" />
        </svg>
      )
    case 'twilio':
      return (
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="w-8 h-8"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          strokeWidth={1.5}
        >
          <title>SMS</title>
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            d="M8.625 9.75a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Zm0 0H8.25m4.125 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Zm0 0H12m4.125 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Zm0 0h-.375M21 12c0 4.556-4.03 8.25-9 8.25a9.764 9.764 0 0 1-2.555-.337A5.972 5.972 0 0 1 5.41 20.97a5.969 5.969 0 0 1-.474-.065 4.48 4.48 0 0 0 .978-2.025c.09-.457-.133-.901-.467-1.226C3.93 16.178 3 14.189 3 12c0-4.556 4.03-8.25 9-8.25s9 3.694 9 8.25Z"
          />
        </svg>
      )
    case 'sheets':
      return (
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="w-8 h-8"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          strokeWidth={1.5}
        >
          <title>Google Sheets</title>
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            d="M9 12h6m-6 3h6m-6 3h6M5.25 4.5h13.5a.75.75 0 0 1 .75.75v13.5a.75.75 0 0 1-.75.75H5.25a.75.75 0 0 1-.75-.75V5.25a.75.75 0 0 1 .75-.75Z"
          />
        </svg>
      )
    default:
      return (
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="w-8 h-8"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          strokeWidth={1.5}
        >
          <title>Integration</title>
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            d="M13.19 8.688a4.5 4.5 0 011.242 7.244l-4.5 4.5a4.5 4.5 0 01-6.364-6.364l1.757-1.757m13.35-.622l1.757-1.757a4.5 4.5 0 00-6.364-6.364l-4.5 4.5a4.5 4.5 0 001.242 7.244"
          />
        </svg>
      )
  }
}

export function ProviderCard({
  providerKey,
  label,
  description,
  enabled,
  authType,
  soon,
  integration,
  planCode,
}: Readonly<ProviderCardProps>) {
  return (
    <div
      className={`relative flex flex-col gap-4 rounded-2xl border border-border bg-card p-5 transition-shadow ${
        enabled ? 'hover:shadow-md' : 'opacity-60'
      }`}
    >
      {/* Coming soon badge */}
      {soon && (
        <span className="absolute top-3 right-3 text-[10px] font-medium bg-muted text-muted-foreground px-2 py-0.5 rounded-full">
          Coming soon
        </span>
      )}

      {/* Header: icon + status */}
      <div className="flex items-start gap-3">
        <div className="w-12 h-12 rounded-xl bg-muted flex items-center justify-center text-muted-foreground flex-shrink-0">
          <ProviderIcon providerKey={providerKey} />
        </div>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="text-sm font-semibold text-foreground">{label}</span>
            {integration && (
              <span className="text-[10px] font-medium px-1.5 py-0.5 rounded-full bg-green-500/15 text-green-600 dark:text-green-400">
                Connected
              </span>
            )}
          </div>
          <p className="text-xs text-muted-foreground mt-0.5 line-clamp-2">{description}</p>
        </div>
      </div>

      {/* Action */}
      {enabled && (
        <ConnectButton
          providerKey={providerKey}
          authType={authType}
          integration={integration}
          label={label}
          planCode={planCode}
        />
      )}
    </div>
  )
}
