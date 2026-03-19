import { cn } from "@/lib/utils";

const paddings = {
  sm: "p-5",
  md: "p-8 md:p-10",
  lg: "p-10 md:p-12",
} as const;

interface CardProps {
  hover?: boolean;
  padding?: keyof typeof paddings;
  className?: string;
  children: React.ReactNode;
}

export function Card({
  hover = true,
  padding = "md",
  className,
  children,
}: CardProps) {
  return (
    <div
      className={cn(
        "rounded-xl border border-border bg-bg-card",
        paddings[padding],
        hover &&
          "snake-border transition-all duration-200 hover:border-transparent hover:-translate-y-0.5",
        className,
      )}
    >
      {children}
    </div>
  );
}
