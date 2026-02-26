import { cn } from "@/lib/utils";

type ButtonBase = {
  variant?: "primary" | "secondary";
  size?: "sm" | "lg";
};

type ButtonAsButton = ButtonBase &
  Omit<React.ButtonHTMLAttributes<HTMLButtonElement>, "href"> & {
    href?: never;
  };

type ButtonAsLink = ButtonBase &
  Omit<React.AnchorHTMLAttributes<HTMLAnchorElement>, "href"> & {
    href: string;
  };

type ButtonProps = ButtonAsButton | ButtonAsLink;

export function Button({
  variant = "primary",
  size = "lg",
  className,
  children,
  href,
  ...props
}: ButtonProps) {
  const base =
    "inline-flex items-center justify-center font-medium rounded-lg transition-all duration-200 cursor-pointer whitespace-nowrap";

  const variants = {
    primary:
      "bg-purple text-white hover:brightness-110 hover:shadow-[0_0_24px_rgba(99,102,241,0.35)]",
    secondary:
      "border border-border-light text-text hover:border-purple/50 hover:text-white bg-transparent",
  };

  const sizes = {
    sm: "px-5 py-2.5 text-[15px] gap-1.5",
    lg: "px-6 py-3 text-[16px] gap-2",
  };

  const classes = cn(base, variants[variant], sizes[size], className);

  if (href) {
    return (
      <a
        href={href}
        className={classes}
        {...(props as React.AnchorHTMLAttributes<HTMLAnchorElement>)}
      >
        {children}
      </a>
    );
  }

  return (
    <button
      className={classes}
      {...(props as React.ButtonHTMLAttributes<HTMLButtonElement>)}
    >
      {children}
    </button>
  );
}
