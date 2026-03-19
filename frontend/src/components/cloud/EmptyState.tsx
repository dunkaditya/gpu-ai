import Link from "next/link";

interface EmptyStateProps {
  icon: React.ReactNode;
  title: string;
  description?: React.ReactNode;
  action?: { label: string; href?: string; onClick?: () => void };
}

export function EmptyState({ icon, title, description, action }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center py-16">
      {/* Icon container */}
      <div className="flex items-center justify-center w-10 h-10 rounded-lg bg-bg-card-hover text-text-dim">
        {icon}
      </div>

      {/* Title */}
      <p className="mt-4 type-ui-sm text-text-muted">{title}</p>

      {/* Description */}
      {description && (
        <p className="mt-1 type-ui-2xs text-text-dim">{description}</p>
      )}

      {/* Action */}
      {action && (
        <div className="mt-4">
          {action.href ? (
            <Link href={action.href} className="btn-primary">
              {action.label}
            </Link>
          ) : (
            <button onClick={action.onClick} className="btn-primary">
              {action.label}
            </button>
          )}
        </div>
      )}
    </div>
  );
}
