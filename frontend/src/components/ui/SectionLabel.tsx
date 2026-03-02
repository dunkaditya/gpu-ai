import { cn } from "@/lib/utils";

export function SectionLabel({
  children,
  className,
}: {
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <span
      className={cn(
        "type-caption font-medium uppercase tracking-[0.1em] text-purple-light",
        className,
      )}
    >
      {children}
    </span>
  );
}
