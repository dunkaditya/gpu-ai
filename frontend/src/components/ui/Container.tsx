import { cn } from "@/lib/utils";

export function Container({
  className,
  children,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div className={cn("mx-auto max-w-[1200px] px-6", className)} {...props}>
      {children}
    </div>
  );
}
