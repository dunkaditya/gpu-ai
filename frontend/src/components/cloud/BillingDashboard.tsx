"use client";

import { useState, useEffect } from "react";
import useSWR from "swr";
import { cn } from "@/lib/utils";
import { fetcher, purchaseCredits, redeemCreditCode, updateAutoPay } from "@/lib/api";
import { EmptyState } from "@/components/cloud/EmptyState";
import type {
  BalanceResponse,
  TransactionsListResponse,
  TransactionResponse,
  UsageResponse,
} from "@/lib/types";

/* ── Helpers ── */

function formatCents(cents: number): string {
  return (cents / 100).toFixed(2);
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function formatDuration(seconds?: number | null): string {
  if (!seconds) return "--";
  const hours = Math.floor(seconds / 3600);
  const mins = Math.floor((seconds % 3600) / 60);
  if (hours > 0) return `${hours}h ${mins}m`;
  return `${mins}m`;
}

const TX_TYPE_LABELS: Record<string, { label: string; color: string }> = {
  credit_purchase: { label: "Purchase", color: "text-green bg-green-dim" },
  auto_pay: { label: "Auto-Pay", color: "text-purple-light bg-purple-dim" },
  credit_code: { label: "Code", color: "text-green bg-green-dim" },
  usage_deduction: { label: "Usage", color: "text-text-muted bg-bg-card-hover" },
  adjustment: { label: "Adjust", color: "text-text-muted bg-bg-card-hover" },
};

const CREDIT_PRESETS = [
  { label: "$10", cents: 1000 },
  { label: "$25", cents: 2500 },
  { label: "$50", cents: 5000 },
  { label: "$100", cents: 10000 },
  { label: "$250", cents: 25000 },
  { label: "$500", cents: 50000 },
];

/* ── Tab buttons ── */

type Tab = "credits" | "usage";

/* ── Main Component ── */

export function BillingDashboard() {
  const [tab, setTab] = useState<Tab>("credits");

  // Payment success/cancelled detection
  const [paymentMessage, setPaymentMessage] = useState<{
    type: "success" | "cancelled";
    text: string;
  } | null>(null);

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    if (params.get("payment") === "success") {
      setPaymentMessage({ type: "success", text: "Payment successful — credits added to your balance." });
      window.history.replaceState({}, "", window.location.pathname);
    } else if (params.get("payment") === "cancelled") {
      setPaymentMessage({ type: "cancelled", text: "Payment was cancelled." });
      window.history.replaceState({}, "", window.location.pathname);
    }
  }, []);

  return (
    <div className="space-y-6">
      {/* Payment toast */}
      {paymentMessage && (
        <div
          className={cn(
            "rounded-[10px] border px-4 py-3 type-ui-sm flex items-center justify-between",
            paymentMessage.type === "success"
              ? "border-green/30 bg-green-dim text-green"
              : "border-border bg-bg-card text-text-muted"
          )}
        >
          <span>{paymentMessage.text}</span>
          <button
            onClick={() => setPaymentMessage(null)}
            className="text-text-dim hover:text-text ml-4"
          >
            ✕
          </button>
        </div>
      )}

      {/* Tab switcher */}
      <div className="flex rounded-lg border border-border overflow-hidden w-fit">
        {(["credits", "usage"] as Tab[]).map((t) => (
          <button
            key={t}
            onClick={() => setTab(t)}
            className={cn(
              "px-4 py-2 type-ui-xs font-medium transition-colors capitalize",
              tab === t
                ? "bg-bg-card-hover text-text"
                : "text-text-muted hover:text-text hover:bg-bg-card"
            )}
          >
            {t === "credits" ? "Credits & Balance" : "Usage Sessions"}
          </button>
        ))}
      </div>

      {tab === "credits" ? <CreditsTab /> : <UsageTab />}
    </div>
  );
}

/* ── Credits Tab ── */

function CreditsTab() {
  const { data: balance, mutate: mutateBalance } = useSWR<BalanceResponse>(
    "/api/v1/billing/balance",
    fetcher,
    { refreshInterval: 15000 }
  );

  const { data: txData, mutate: mutateTx } = useSWR<TransactionsListResponse>(
    "/api/v1/billing/transactions?limit=25",
    fetcher,
    { refreshInterval: 30000 }
  );

  // Estimate spend rate from recent usage_deduction transactions
  const spendRate = getSpendRate(txData?.transactions);

  return (
    <div className="space-y-6">
      {/* Balance + Spend Rate */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {/* Balance */}
        <div className="bg-bg-card border border-border rounded-[10px] p-5">
          <p className="type-ui-xs text-text-dim font-medium uppercase tracking-wider mb-2">
            Credit Balance
          </p>
          <p className="type-h2 font-mono text-text">
            <span className="text-text-muted">$</span>
            {balance ? formatCents(balance.balance_cents) : "-.--"}
          </p>
          {balance && balance.balance_cents <= 0 && (
            <p className="type-ui-2xs text-red-400 mt-1">
              Balance depleted — add credits to launch instances
            </p>
          )}
        </div>

        {/* Spend Rate */}
        <div className="bg-bg-card border border-border rounded-[10px] p-5">
          <p className="type-ui-xs text-text-dim font-medium uppercase tracking-wider mb-2">
            Burn Rate
          </p>
          {spendRate !== null ? (
            <>
              <p className="type-h2 font-mono text-text">
                <span className="text-text-muted">$</span>
                {spendRate.toFixed(2)}
                <span className="type-body-sm text-text-dim">/hr</span>
              </p>
              {balance && spendRate > 0 && (
                <p className="type-ui-2xs text-text-dim mt-1">
                  ~{Math.floor(balance.balance_cents / 100 / spendRate)}h remaining
                </p>
              )}
            </>
          ) : (
            <p className="type-ui-sm text-text-dim mt-1">No active usage</p>
          )}
        </div>

        {/* Auto-Pay Status */}
        <div className="bg-bg-card border border-border rounded-[10px] p-5">
          <p className="type-ui-xs text-text-dim font-medium uppercase tracking-wider mb-2">
            Auto-Pay
          </p>
          {balance?.auto_pay_enabled ? (
            <>
              <p className="type-h3 font-mono text-green">Enabled</p>
              <p className="type-ui-2xs text-text-dim mt-1">
                Charges ${formatCents(balance.auto_pay_amount_cents)} when balance drops below $
                {formatCents(balance.auto_pay_threshold_cents)}
              </p>
            </>
          ) : (
            <p className="type-ui-sm text-text-dim mt-1">
              Disabled — configure below
            </p>
          )}
        </div>
      </div>

      {/* Add Credits */}
      <AddCreditsSection onSuccess={() => { mutateBalance(); mutateTx(); }} />

      {/* Redeem Code */}
      <RedeemCodeSection onSuccess={() => { mutateBalance(); mutateTx(); }} />

      {/* Auto-Pay Configuration */}
      <AutoPaySection balance={balance} onSuccess={() => mutateBalance()} />

      {/* Transaction History */}
      <TransactionHistorySection data={txData} />
    </div>
  );
}

/* ── Add Credits Section ── */

function AddCreditsSection({ onSuccess }: { onSuccess: () => void }) {
  const [selectedPreset, setSelectedPreset] = useState<number | null>(null);
  const [customAmount, setCustomAmount] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const amountCents =
    selectedPreset ?? (customAmount ? Math.round(parseFloat(customAmount) * 100) : 0);

  async function handlePurchase() {
    if (amountCents < 500) {
      setError("Minimum purchase is $5.00");
      return;
    }
    setLoading(true);
    setError("");
    try {
      const baseUrl = window.location.origin + "/cloud/billing";
      const result = await purchaseCredits(
        amountCents,
        baseUrl + "?payment=success",
        baseUrl + "?payment=cancelled"
      );
      window.location.href = result.checkout_url;
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "Payment failed");
      setLoading(false);
    }
  }

  return (
    <div className="bg-bg-card border border-border rounded-[10px] p-5">
      <h3 className="type-ui-sm text-text font-medium mb-4">Add Credits</h3>

      {/* Preset buttons */}
      <div className="flex flex-wrap gap-2 mb-4">
        {CREDIT_PRESETS.map((p) => (
          <button
            key={p.cents}
            onClick={() => { setSelectedPreset(p.cents); setCustomAmount(""); setError(""); }}
            className={cn(
              "px-4 py-2 rounded-md type-ui-sm font-mono border transition-colors",
              selectedPreset === p.cents
                ? "border-purple bg-purple-dim text-purple-light"
                : "border-border text-text-muted hover:text-text hover:border-border-light"
            )}
          >
            {p.label}
          </button>
        ))}
      </div>

      {/* Custom amount + pay button */}
      <div className="flex items-center gap-3">
        <div className="relative flex-1 max-w-[200px]">
          <span className="absolute left-3 top-1/2 -translate-y-1/2 text-text-dim type-ui-sm pointer-events-none font-mono">$</span>
          <input
            type="number"
            min="5"
            step="1"
            placeholder="Other"
            value={customAmount}
            onChange={(e) => {
              setCustomAmount(e.target.value);
              setSelectedPreset(null);
              setError("");
            }}
            className="w-full pl-[42px] pr-3 py-2 bg-bg border border-border rounded-md type-ui-sm
                       text-text placeholder:text-text-dim focus:outline-none focus:border-purple
                       font-mono [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none"
          />
        </div>
        <button
          onClick={handlePurchase}
          disabled={amountCents < 500 || loading}
          className="btn-primary px-5 py-2 disabled:opacity-40"
        >
          {loading ? "Redirecting..." : "Pay with card"}
        </button>
      </div>

      {error && <p className="type-ui-2xs text-red-400 mt-2">{error}</p>}
    </div>
  );
}

/* ── Redeem Code Section ── */

function RedeemCodeSection({ onSuccess }: { onSuccess: () => void }) {
  const [code, setCode] = useState("");
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState<{ type: "success" | "error"; text: string } | null>(null);

  async function handleRedeem() {
    if (!code.trim()) return;
    setLoading(true);
    setMessage(null);
    try {
      const result = await redeemCreditCode(code.trim().toUpperCase());
      setMessage({
        type: "success",
        text: `Redeemed $${result.amount_dollars.toFixed(2)} — new balance: $${formatCents(result.new_balance_cents)}`,
      });
      setCode("");
      onSuccess();
    } catch (e: unknown) {
      setMessage({
        type: "error",
        text: e instanceof Error ? e.message : "Failed to redeem code",
      });
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="bg-bg-card border border-border rounded-[10px] p-5">
      <h3 className="type-ui-sm text-text font-medium mb-3">Redeem Credit Code</h3>
      <div className="flex items-center gap-3">
        <input
          type="text"
          value={code}
          onChange={(e) => { setCode(e.target.value); setMessage(null); }}
          placeholder="GPU-XXXX-XXXX"
          className="flex-1 max-w-[240px] px-3 py-2 bg-bg border border-border rounded-md type-ui-sm
                     text-text placeholder:text-text-dim focus:outline-none focus:border-purple
                     font-mono uppercase tracking-wider"
        />
        <button
          onClick={handleRedeem}
          disabled={!code.trim() || loading}
          className="btn-secondary px-4 py-2 disabled:opacity-40"
        >
          {loading ? "Redeeming..." : "Redeem Code"}
        </button>
      </div>
      {message && (
        <p
          className={cn(
            "type-ui-2xs mt-2",
            message.type === "success" ? "text-green" : "text-red-400"
          )}
        >
          {message.text}
        </p>
      )}
    </div>
  );
}

/* ── Auto-Pay Section ── */

function AutoPaySection({
  balance,
  onSuccess,
}: {
  balance?: BalanceResponse | null;
  onSuccess: () => void;
}) {
  const [enabled, setEnabled] = useState(false);
  const [threshold, setThreshold] = useState("10");
  const [amount, setAmount] = useState("25");
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState<{ type: "success" | "error"; text: string } | null>(null);

  // Sync from server
  useEffect(() => {
    if (balance) {
      setEnabled(balance.auto_pay_enabled);
      if (balance.auto_pay_threshold_cents > 0)
        setThreshold((balance.auto_pay_threshold_cents / 100).toString());
      if (balance.auto_pay_amount_cents > 0)
        setAmount((balance.auto_pay_amount_cents / 100).toString());
    }
  }, [balance]);

  async function handleSave() {
    setLoading(true);
    setMessage(null);
    try {
      await updateAutoPay(
        enabled,
        Math.round(parseFloat(threshold || "0") * 100),
        Math.round(parseFloat(amount || "0") * 100)
      );
      setMessage({ type: "success", text: "Auto-pay settings saved." });
      onSuccess();
    } catch (e: unknown) {
      setMessage({ type: "error", text: e instanceof Error ? e.message : "Failed to save" });
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="bg-bg-card border border-border rounded-[10px] p-5">
      <div className="flex items-center justify-between mb-4">
        <h3 className="type-ui-sm text-text font-medium">Auto-Pay</h3>
        <button
          onClick={() => { setEnabled(!enabled); setMessage(null); }}
          className={cn(
            "relative w-10 h-5 rounded-full transition-colors",
            enabled ? "bg-purple" : "bg-border"
          )}
        >
          <span
            className={cn(
              "absolute top-0.5 w-4 h-4 rounded-full bg-white transition-transform",
              enabled ? "left-[22px]" : "left-0.5"
            )}
          />
        </button>
      </div>

      {enabled && (
        <div className="flex flex-wrap items-end gap-4">
          <div>
            <label className="type-ui-2xs text-text-dim block mb-1">When balance drops below</label>
            <div className="relative">
              <span className="absolute left-3 top-1/2 -translate-y-1/2 text-text-dim type-ui-sm">$</span>
              <input
                type="number"
                min="0"
                step="5"
                value={threshold}
                onChange={(e) => setThreshold(e.target.value)}
                className="w-[120px] pl-7 pr-3 py-2 bg-bg border border-border rounded-md type-ui-sm
                           text-text font-mono focus:outline-none focus:border-purple
                           [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none"
              />
            </div>
          </div>
          <div>
            <label className="type-ui-2xs text-text-dim block mb-1">Charge amount</label>
            <div className="relative">
              <span className="absolute left-3 top-1/2 -translate-y-1/2 text-text-dim type-ui-sm">$</span>
              <input
                type="number"
                min="5"
                step="5"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                className="w-[120px] pl-7 pr-3 py-2 bg-bg border border-border rounded-md type-ui-sm
                           text-text font-mono focus:outline-none focus:border-purple
                           [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none"
              />
            </div>
          </div>
          <button
            onClick={handleSave}
            disabled={loading}
            className="btn-secondary px-4 py-2 disabled:opacity-40"
          >
            {loading ? "Saving..." : "Save"}
          </button>
        </div>
      )}

      {message && (
        <p
          className={cn(
            "type-ui-2xs mt-2",
            message.type === "success" ? "text-green" : "text-red-400"
          )}
        >
          {message.text}
        </p>
      )}
    </div>
  );
}

/* ── Transaction History ── */

function TransactionHistorySection({ data }: { data?: TransactionsListResponse | null }) {
  const transactions = data?.transactions ?? [];

  return (
    <div className="rounded-[10px] border border-border bg-bg-card/50 overflow-hidden">
      <div className="px-5 py-3 border-b border-border">
        <h3 className="type-ui-sm text-text font-medium">Transaction History</h3>
      </div>
      <div className="overflow-x-auto">
        <table className="w-full">
          <thead>
            <tr className="border-b border-border">
              <th className="type-ui-2xs text-left text-text-dim font-medium uppercase tracking-wider px-4 py-3">
                Date
              </th>
              <th className="type-ui-2xs text-left text-text-dim font-medium uppercase tracking-wider px-4 py-3">
                Type
              </th>
              <th className="type-ui-2xs text-left text-text-dim font-medium uppercase tracking-wider px-4 py-3">
                Description
              </th>
              <th className="type-ui-2xs text-right text-text-dim font-medium uppercase tracking-wider px-4 py-3">
                Amount
              </th>
              <th className="type-ui-2xs text-right text-text-dim font-medium uppercase tracking-wider px-4 py-3">
                Balance
              </th>
            </tr>
          </thead>
          <tbody>
            {transactions.length === 0 ? (
              <tr>
                <td colSpan={5}>
                  <EmptyState
                    icon={
                      <svg width="20" height="20" viewBox="0 0 16 16" fill="none">
                        <rect x="1" y="3" width="14" height="10" rx="1.5" stroke="currentColor" strokeWidth="1.5" />
                        <path d="M1 6h14" stroke="currentColor" strokeWidth="1.5" />
                        <rect x="3" y="8.5" width="4" height="2" rx="0.5" stroke="currentColor" strokeWidth="1" />
                      </svg>
                    }
                    title="No transactions yet"
                    description="Add credits or launch an instance to start tracking."
                  />
                </td>
              </tr>
            ) : (
              transactions.map((tx) => {
                const info = TX_TYPE_LABELS[tx.type] ?? {
                  label: tx.type,
                  color: "text-text-muted bg-bg-card-hover",
                };
                return (
                  <tr
                    key={tx.id}
                    className="border-b border-border/50 hover:bg-bg-card transition-colors"
                  >
                    <td className="px-4 py-3">
                      <span className="type-ui-xs text-text-muted">
                        {formatDate(tx.created_at)}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <span
                        className={cn(
                          "type-ui-2xs font-medium rounded-full px-2 py-0.5",
                          info.color
                        )}
                      >
                        {info.label}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <span className="type-ui-sm text-text-muted">
                        {tx.description || "--"}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-right">
                      <span
                        className={cn(
                          "type-ui-sm font-mono",
                          tx.amount_cents >= 0 ? "text-green" : "text-text"
                        )}
                      >
                        {tx.amount_cents >= 0 ? "+" : ""}${formatCents(Math.abs(tx.amount_cents))}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-right">
                      <span className="type-ui-sm font-mono text-text-muted">
                        ${formatCents(tx.balance_after_cents)}
                      </span>
                    </td>
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}

/* ── Usage Sessions Tab (existing billing sessions) ── */

function UsageTab() {
  const [period, setPeriod] = useState("current_month");
  const { data, error, isLoading, mutate } = useSWR<UsageResponse>(
    `/api/v1/billing/usage?period=${period}`,
    fetcher,
    { refreshInterval: 60000 }
  );

  const periods = [
    { label: "Current Month", value: "current_month" },
    { label: "Last 30 Days", value: "last_30d" },
  ] as const;

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-center">
        <p className="type-ui-sm text-red-400">Failed to load usage data</p>
        <button
          onClick={() => mutate()}
          className="mt-3 type-ui-xs text-text-muted hover:text-text transition-colors"
        >
          Retry
        </button>
      </div>
    );
  }

  const sessions = data?.sessions ?? [];
  const activeSessions = sessions.filter((s) => s.is_active);

  return (
    <div className="space-y-6">
      {/* Summary */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div className="bg-bg-card border border-border rounded-[10px] p-5">
          <p className="type-ui-xs text-text-dim font-medium uppercase tracking-wider mb-2">
            Total Cost
          </p>
          {isLoading ? (
            <div className="h-10 bg-bg-card-hover rounded animate-pulse w-32" />
          ) : (
            <p className="type-h2 font-mono text-text">
              <span className="text-text-muted">$</span>
              {(data?.total_cost ?? 0).toFixed(2)}
            </p>
          )}
        </div>
        <div className="bg-bg-card border border-border rounded-[10px] p-5">
          <p className="type-ui-xs text-text-dim font-medium uppercase tracking-wider mb-2">
            Active Sessions
          </p>
          {isLoading ? (
            <div className="h-10 bg-bg-card-hover rounded animate-pulse w-12" />
          ) : (
            <p className="type-h2 font-mono text-green">{activeSessions.length}</p>
          )}
        </div>
      </div>

      {/* Period Selector */}
      <div className="flex rounded-lg border border-border overflow-hidden w-fit">
        {periods.map((p) => (
          <button
            key={p.value}
            onClick={() => setPeriod(p.value)}
            className={cn(
              "px-4 py-2 type-ui-xs font-medium transition-colors",
              period === p.value
                ? "bg-bg-card-hover text-text"
                : "text-text-muted hover:text-text hover:bg-bg-card"
            )}
          >
            {p.label}
          </button>
        ))}
      </div>

      {/* Sessions Table */}
      <div className="rounded-[10px] border border-border bg-bg-card/50 overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-border">
                <th className="type-ui-2xs text-left text-text-dim font-medium uppercase tracking-wider px-4 py-3">GPU</th>
                <th className="type-ui-2xs text-left text-text-dim font-medium uppercase tracking-wider px-4 py-3">Count</th>
                <th className="type-ui-2xs text-left text-text-dim font-medium uppercase tracking-wider px-4 py-3">Rate</th>
                <th className="type-ui-2xs text-left text-text-dim font-medium uppercase tracking-wider px-4 py-3">Started</th>
                <th className="type-ui-2xs text-left text-text-dim font-medium uppercase tracking-wider px-4 py-3">Ended</th>
                <th className="type-ui-2xs text-left text-text-dim font-medium uppercase tracking-wider px-4 py-3">Duration</th>
                <th className="type-ui-2xs text-right text-text-dim font-medium uppercase tracking-wider px-4 py-3">Cost</th>
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                Array.from({ length: 4 }).map((_, i) => (
                  <tr key={i} className="border-b border-border/50">
                    {Array.from({ length: 7 }).map((_, j) => (
                      <td key={j} className="px-4 py-3">
                        <div className="h-4 bg-bg-card-hover rounded animate-pulse w-16" />
                      </td>
                    ))}
                  </tr>
                ))
              ) : sessions.length === 0 ? (
                <tr>
                  <td colSpan={7}>
                    <EmptyState
                      icon={
                        <svg width="20" height="20" viewBox="0 0 16 16" fill="none">
                          <rect x="1" y="3" width="14" height="10" rx="1.5" stroke="currentColor" strokeWidth="1.5" />
                          <path d="M1 6h14" stroke="currentColor" strokeWidth="1.5" />
                          <rect x="3" y="8.5" width="4" height="2" rx="0.5" stroke="currentColor" strokeWidth="1" />
                        </svg>
                      }
                      title="No billing sessions in this period"
                      description="Launch an instance to start tracking usage."
                    />
                  </td>
                </tr>
              ) : (
                sessions.map((session) => (
                  <tr
                    key={session.id}
                    className="border-b border-border/50 hover:bg-bg-card transition-colors"
                  >
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-2">
                        {session.is_active && (
                          <span className="h-1.5 w-1.5 rounded-full bg-green animate-pulse-dot shrink-0" />
                        )}
                        <span className="type-ui-sm text-text font-medium">{session.gpu_type}</span>
                      </div>
                    </td>
                    <td className="px-4 py-3">
                      <span className="type-ui-sm text-text font-mono">x{session.gpu_count}</span>
                    </td>
                    <td className="px-4 py-3">
                      <span className="type-ui-sm text-text-muted font-mono">
                        ${session.price_per_hour.toFixed(2)}/hr
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <span className="type-ui-xs text-text-muted">{formatDate(session.started_at)}</span>
                    </td>
                    <td className="px-4 py-3">
                      {session.ended_at ? (
                        <span className="type-ui-xs text-text-muted">{formatDate(session.ended_at)}</span>
                      ) : (
                        <span className="type-ui-xs inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 bg-green-dim text-green font-medium">
                          <span className="h-1.5 w-1.5 rounded-full bg-green" />
                          Active
                        </span>
                      )}
                    </td>
                    <td className="px-4 py-3">
                      <span className="type-ui-sm text-text-muted font-mono">
                        {formatDuration(session.duration_seconds)}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-right">
                      <span className="type-ui-sm text-text font-mono">
                        ${(session.total_cost ?? session.estimated_cost ?? 0).toFixed(2)}
                      </span>
                      {session.is_active && session.estimated_cost != null && (
                        <p className="type-ui-2xs text-text-dim">est.</p>
                      )}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}

/* ── Helpers ── */

/** Estimate $/hr spend rate from recent usage_deduction transactions. */
function getSpendRate(transactions?: TransactionResponse[]): number | null {
  if (!transactions) return null;
  const deductions = transactions.filter((t) => t.type === "usage_deduction");
  if (deductions.length < 2) return null;

  // Use the last few deductions to estimate rate
  const recent = deductions.slice(0, Math.min(5, deductions.length));
  const totalCents = recent.reduce((sum, t) => sum + Math.abs(t.amount_cents), 0);

  const newest = new Date(recent[0].created_at).getTime();
  const oldest = new Date(recent[recent.length - 1].created_at).getTime();
  const hoursSpan = (newest - oldest) / (1000 * 60 * 60);

  if (hoursSpan <= 0) return null;
  return totalCents / 100 / hoursSpan;
}
