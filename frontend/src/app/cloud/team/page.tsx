"use client";

import { OrganizationProfile, CreateOrganization, useOrganization } from "@clerk/nextjs";

const hasClerk = !!process.env.NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY;

function TeamContent() {
  const { organization, isLoaded } = useOrganization();

  if (!isLoaded) return null;

  if (!organization) {
    return (
      <div className="rounded-lg border border-border bg-bg-card/50 overflow-hidden">
        <div className="flex flex-col items-center justify-center py-16 text-center px-4">
          <div className="w-14 h-14 rounded-xl bg-purple-dim flex items-center justify-center mb-5">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <circle cx="9" cy="7" r="3.5" stroke="var(--color-purple-light)" strokeWidth="2" />
              <path d="M2 20c0-3.5 3-6.5 7-6.5s7 3 7 6.5" stroke="var(--color-purple-light)" strokeWidth="2" strokeLinecap="round" />
              <circle cx="17" cy="8" r="2.5" stroke="var(--color-purple-light)" strokeWidth="1.5" />
              <path d="M17 13c2.5 0 5 1.5 5 4.5" stroke="var(--color-purple-light)" strokeWidth="1.5" strokeLinecap="round" />
            </svg>
          </div>
          <h2 className="type-h5 font-sans text-text">Create an Organization</h2>
          <p className="mt-2 type-ui-sm text-text-muted max-w-sm mb-8">
            You&apos;re on a personal account. Create an organization to invite
            team members, manage roles, and share GPU instances.
          </p>
          <CreateOrganization
            afterCreateOrganizationUrl="/cloud/team"
            appearance={{
              elements: {
                rootBox: "w-full max-w-md",
                cardBox: "w-full shadow-none",
                card: "w-full bg-transparent shadow-none border-0",
                headerTitle: "text-text",
                headerSubtitle: "text-text-muted",
                formFieldInput:
                  "bg-bg-card border border-border text-text",
                formButtonPrimary:
                  "bg-purple hover:bg-purple-light text-white",
              },
            }}
          />
        </div>
      </div>
    );
  }

  return (
    <OrganizationProfile
      routing="hash"
      appearance={{
        elements: {
          rootBox: "w-full",
          cardBox: "w-full shadow-none",
          card: "w-full bg-transparent shadow-none border border-border rounded-lg",
          navbar: "bg-bg-alt border-r border-border",
          navbarButton: "text-text-muted hover:text-text",
          navbarButtonActive: "text-text",
          pageScrollBox: "bg-transparent",
          headerTitle: "text-text",
          headerSubtitle: "text-text-muted",
          profileSectionTitle: "text-text",
          profileSectionContent: "text-text-muted",
          membersPageInviteButton:
            "bg-purple hover:bg-purple-light text-white",
          tableHead: "text-text-dim",
          tableBody: "text-text",
        },
      }}
    />
  );
}

export default function TeamPage() {
  return (
    <div className="space-y-6">
      <h1 className="type-h3 text-text">Team</h1>
      {hasClerk ? (
        <TeamContent />
      ) : (
        <div className="rounded-lg border border-border bg-bg-card/50 p-8 text-center">
          <p className="type-ui-sm text-text-muted">
            Team management requires Clerk to be configured.
          </p>
        </div>
      )}
    </div>
  );
}
