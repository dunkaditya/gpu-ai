import { cn } from "@/lib/utils";

interface ButtonProps extends React.AnchorHTMLAttributes<HTMLAnchorElement> {
  variant?: "primary" | "secondary";
  size?: "sm" | "md" | "lg";
}

export function Button({
  variant = "primary",
  size = "md",
  className,
  children,
  ...props
}: ButtonProps) {
  return (
    <a
      className={cn(
        "inline-flex items-center justify-center rounded-full font-medium transition-colors",
        variant === "primary" && "bg-white text-black hover:bg-gray-200",
        variant === "secondary" &&
          "border border-white/20 text-white hover:border-white/40 hover:bg-white/5",
        size === "sm" && "px-4 py-1.5 text-sm",
        size === "md" && "px-6 py-2 text-sm",
        size === "lg" && "px-8 py-3 text-base",
        className
      )}
      {...props}
    >
      {children}
    </a>
  );
}
