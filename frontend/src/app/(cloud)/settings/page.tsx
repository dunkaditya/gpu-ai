import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Settings",
};

export default function SettingsPage() {
  return (
    <div className="space-y-4">
      <h1 className="type-h3 text-text">Settings</h1>
      <p className="type-body text-text-muted">
        Settings page coming soon.
      </p>
    </div>
  );
}
