export function UpgradeBanner() {
  return (
    <div className="rounded-2xl bg-sidebar-accent px-8 py-6 flex items-center justify-between gap-4 overflow-hidden">
      <div className="flex-1">
        <p className="text-sm text-muted-foreground font-medium">Don&apos;t miss what matters next</p>
        <h2 className="text-xl font-bold text-foreground mt-1 leading-snug">
          More alerts, deeper context, and faster intelligence,
          <br />
          all with a simple upgrade.
        </h2>
        <button
          type="button"
          className="mt-4 inline-flex items-center gap-2 bg-primary hover:bg-primary/90 text-primary-foreground text-sm font-medium px-5 py-2.5 rounded-full transition-colors"
        >
          Check pricing &rarr;
        </button>
      </div>
      <div className="hidden sm:block flex-shrink-0 w-32 h-32 relative">
        {/* Decorative illustration placeholder */}
        <div className="w-full h-full rounded-full bg-primary/15 flex items-center justify-center">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            className="w-16 h-16 text-primary opacity-60"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            strokeWidth={1.5}
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              d="M9.75 3.104v5.714a2.25 2.25 0 01-.659 1.591L5 14.5M9.75 3.104c-.251.023-.501.05-.75.082m.75-.082a24.301 24.301 0 014.5 0m0 0v5.714c0 .597.237 1.17.659 1.591L19.8 15.3M14.25 3.104c.251.023.501.05.75.082M19.8 15.3l-1.57.393A9.065 9.065 0 0112 15a9.065 9.065 0 00-6.23-.693L5 14.5m14.8.8l1.402 1.402c1 1 .03 2.698-1.382 2.698H4.18c-1.412 0-2.38-1.698-1.381-2.698L4.2 15.3"
            />
          </svg>
        </div>
      </div>
    </div>
  )
}
