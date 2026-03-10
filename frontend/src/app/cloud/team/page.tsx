export default function TeamPage() {
  return (
    <div className="space-y-6">
      <h1 className="type-h3 text-text">Team</h1>

      <div className="rounded-lg border border-border bg-bg-card/50 overflow-hidden">
        <div className="flex flex-col items-center justify-center py-20 text-center">
          {/* Icon */}
          <div className="w-14 h-14 rounded-xl bg-purple-dim flex items-center justify-center mb-5">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <circle cx="9" cy="7" r="3.5" stroke="var(--color-purple-light)" strokeWidth="2" />
              <path d="M2 20c0-3.5 3-6.5 7-6.5s7 3 7 6.5" stroke="var(--color-purple-light)" strokeWidth="2" strokeLinecap="round" />
              <circle cx="17" cy="8" r="2.5" stroke="var(--color-purple-light)" strokeWidth="1.5" />
              <path d="M17 13c2.5 0 5 1.5 5 4.5" stroke="var(--color-purple-light)" strokeWidth="1.5" strokeLinecap="round" />
            </svg>
          </div>

          <h2 className="type-h5 font-sans text-text">Coming Soon</h2>
          <p className="mt-2 type-ui-sm text-text-muted max-w-sm">
            Manage team members and roles. Invite collaborators, assign permissions,
            and control organization access.
          </p>

          <span className="mt-6 inline-flex items-center gap-2 px-3 py-1.5 rounded-full bg-purple-dim type-ui-2xs text-purple-light">
            <span className="w-1.5 h-1.5 rounded-full bg-purple-light animate-pulse-dot" />
            In Development
          </span>
        </div>
      </div>
    </div>
  );
}
