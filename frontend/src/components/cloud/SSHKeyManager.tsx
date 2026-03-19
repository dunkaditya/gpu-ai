"use client";

import { useState } from "react";
import Link from "next/link";
import useSWR from "swr";
import { cn } from "@/lib/utils";
import { fetcher, addSSHKey, deleteSSHKey } from "@/lib/api";
import { ConfirmDialog } from "@/components/cloud/ConfirmDialog";
import { EmptyState } from "@/components/cloud/EmptyState";
import type { SSHKeyResponse } from "@/lib/types";

function SkeletonRow() {
  return (
    <div className="flex items-center justify-between px-5 py-4 border-b border-border/50">
      <div className="space-y-2">
        <div className="h-4 bg-bg-card-hover rounded animate-pulse w-32" />
        <div className="h-3 bg-bg-card-hover rounded animate-pulse w-48" />
      </div>
      <div className="h-4 bg-bg-card-hover rounded animate-pulse w-16" />
    </div>
  );
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

function truncateFingerprint(fp: string) {
  if (fp.length <= 24) return fp;
  return fp.slice(0, 12) + "..." + fp.slice(-8);
}

export function SSHKeyManager() {
  const { data, error, isLoading, mutate } = useSWR<{
    ssh_keys: SSHKeyResponse[];
  }>("/api/v1/ssh-keys", fetcher);

  const [showForm, setShowForm] = useState(false);
  const [name, setName] = useState("");
  const [publicKey, setPublicKey] = useState("");
  const [addLoading, setAddLoading] = useState(false);
  const [addError, setAddError] = useState<string | null>(null);
  const [deletingKeyId, setDeletingKeyId] = useState<string | null>(null);
  const [deleteLoading, setDeleteLoading] = useState(false);

  const keys = data?.ssh_keys ?? [];

  async function handleAdd(e: React.FormEvent) {
    e.preventDefault();
    setAddError(null);
    setAddLoading(true);
    try {
      await addSSHKey(name, publicKey);
      await mutate();
      setName("");
      setPublicKey("");
      setShowForm(false);
    } catch (err) {
      setAddError(
        err instanceof Error ? err.message : "Failed to add SSH key"
      );
    } finally {
      setAddLoading(false);
    }
  }

  async function handleDeleteConfirm() {
    if (!deletingKeyId) return;
    setDeleteLoading(true);
    try {
      await deleteSSHKey(deletingKeyId);
      await mutate();
      setDeletingKeyId(null);
    } catch {
      // Error will reflect on next fetch
      setDeletingKeyId(null);
    } finally {
      setDeleteLoading(false);
    }
  }

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-center">
        <p className="type-ui-sm text-red-400">Failed to load SSH keys</p>
        <button
          onClick={() => mutate()}
          className="mt-3 type-ui-xs text-text-muted hover:text-text transition-colors"
        >
          Retry
        </button>
      </div>
    );
  }

  return (
    <div className="space-y-5">
      {/* Add key form */}
      {showForm && (
        <form
          onSubmit={handleAdd}
          className="bg-bg-card border border-border rounded-[10px] p-6 space-y-5"
        >
          <div className="space-y-2">
            <label className="type-ui-xs text-text-muted font-medium uppercase tracking-wider">
              Key Name
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g. work-laptop, macbook-pro"
              required
              className="w-full bg-bg border border-border rounded-lg px-4 py-2.5 type-ui-sm text-text placeholder:text-text-dim focus:outline-none focus:ring-1 focus:ring-border-light focus:border-border-light transition-all"
            />
          </div>
          <div className="space-y-2">
            <label className="type-ui-xs text-text-muted font-medium uppercase tracking-wider">
              Public Key
            </label>
            <textarea
              value={publicKey}
              onChange={(e) => setPublicKey(e.target.value)}
              placeholder="ssh-ed25519 AAAA... user@hostname"
              required
              rows={3}
              className="w-full bg-bg border border-border rounded-lg px-4 py-2.5 type-ui-sm text-text font-mono placeholder:text-text-dim focus:outline-none focus:ring-1 focus:ring-border-light focus:border-border-light transition-all resize-none"
            />
          </div>

          {addError && (
            <div className="bg-red-500/10 border border-red-500/30 rounded-lg px-4 py-3">
              <p className="type-ui-xs text-red-400">{addError}</p>
            </div>
          )}

          <div className="flex gap-3">
            <button
              type="button"
              onClick={() => {
                setShowForm(false);
                setAddError(null);
              }}
              className="btn-secondary"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={addLoading || !name || !publicKey}
              className={cn(
                "btn-primary",
                (addLoading || !name || !publicKey) &&
                  "opacity-50 cursor-not-allowed"
              )}
            >
              {addLoading ? "Adding..." : "Add Key"}
            </button>
          </div>
        </form>
      )}

      {/* Keys list */}
      <div className="rounded-[10px] border border-border bg-bg-card/50 overflow-hidden">
        {/* Section header */}
        <div className="flex items-center justify-between px-5 py-3.5 border-b border-border bg-bg-card/80">
          <span className="type-ui-sm text-text font-medium">
            SSH Keys{!isLoading && ` (${keys.length})`}
          </span>
          {!showForm && (
            <button
              onClick={() => setShowForm(true)}
              className="btn-primary !py-1.5 !px-3.5 !type-ui-xs"
            >
              Add Key
            </button>
          )}
        </div>

        {isLoading ? (
          <>
            <SkeletonRow />
            <SkeletonRow />
            <SkeletonRow />
          </>
        ) : keys.length === 0 ? (
          <EmptyState
            icon={
              <svg
                width="20"
                height="20"
                viewBox="0 0 16 16"
                fill="none"
              >
                <circle
                  cx="6"
                  cy="7"
                  r="3"
                  stroke="currentColor"
                  strokeWidth="1.5"
                />
                <path
                  d="M8.5 9.5L14 15"
                  stroke="currentColor"
                  strokeWidth="1.5"
                  strokeLinecap="round"
                />
                <path
                  d="M12 13L14 11"
                  stroke="currentColor"
                  strokeWidth="1.5"
                  strokeLinecap="round"
                />
              </svg>
            }
            title="No SSH keys"
            description={
              <>
                You need an SSH key to connect to instances.{" "}
                <Link
                  href="/docs/ssh-keys"
                  target="_blank"
                  className="text-purple hover:text-purple-light transition-colors"
                >
                  Learn how to create one
                </Link>
              </>
            }
          />
        ) : (
          <div>
            {/* Header row */}
            <div className="hidden md:grid grid-cols-[1fr_1fr_auto_auto] gap-4 px-5 py-3 border-b border-border/60">
              <span className="type-ui-2xs text-text-dim font-medium uppercase tracking-wider">
                Name
              </span>
              <span className="type-ui-2xs text-text-dim font-medium uppercase tracking-wider">
                Fingerprint
              </span>
              <span className="type-ui-2xs text-text-dim font-medium uppercase tracking-wider">
                Added
              </span>
              <span className="type-ui-2xs text-text-dim font-medium uppercase tracking-wider text-right">
                Actions
              </span>
            </div>

            {keys.map((key) => (
              <div
                key={key.id}
                className="flex flex-col md:grid md:grid-cols-[1fr_1fr_auto_auto] gap-2 md:gap-4 md:items-center px-5 py-4 border-b border-border/30 hover:bg-bg-card/80 transition-colors"
              >
                <div>
                  <p className="type-ui-sm text-text font-medium">{key.name}</p>
                </div>
                <div>
                  <code className="type-ui-2xs text-text-muted font-mono bg-bg/60 px-2 py-0.5 rounded">
                    {truncateFingerprint(key.fingerprint)}
                  </code>
                </div>
                <div>
                  <span className="type-ui-xs text-text-dim">
                    {formatDate(key.created_at)}
                  </span>
                </div>
                <div className="md:text-right">
                  <button
                    onClick={() => setDeletingKeyId(key.id)}
                    className="type-ui-2xs px-2.5 py-1 rounded border border-red-500/30 text-red-400 hover:bg-red-500/10 hover:border-red-500/50 transition-colors font-medium"
                  >
                    Delete
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Delete confirmation dialog */}
      {deletingKeyId && (
        <ConfirmDialog
          title="Delete SSH Key"
          message="Instances using this key will no longer accept connections with it. This action cannot be undone."
          confirmLabel="Delete"
          confirmVariant="danger"
          onConfirm={handleDeleteConfirm}
          onCancel={() => setDeletingKeyId(null)}
          loading={deleteLoading}
        />
      )}
    </div>
  );
}
