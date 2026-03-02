import { cn } from "@/lib/utils";
import { Container } from "./Container";
import { SectionLabel } from "./SectionLabel";

interface SectionProps {
  id?: string;
  border?: boolean;
  className?: string;
  containerClassName?: string;
  children: React.ReactNode;
}

export function Section({
  id,
  border = true,
  className,
  containerClassName,
  children,
}: SectionProps) {
  return (
    <section
      id={id}
      className={cn(
        "py-24 md:py-32",
        border && "border-t border-border",
        className,
      )}
    >
      <Container className={containerClassName}>{children}</Container>
    </section>
  );
}

interface SectionHeaderProps {
  label: string;
  title: string;
  description?: string;
  className?: string;
}

export function SectionHeader({
  label,
  title,
  description,
  className,
}: SectionHeaderProps) {
  return (
    <div className={cn("mb-12 text-center md:mb-16", className)}>
      <SectionLabel>{label}</SectionLabel>
      <h2 className="type-h2 mt-3 font-bold text-white">{title}</h2>
      {description && (
        <p className="type-body-lg mt-3 text-text-muted">{description}</p>
      )}
    </div>
  );
}
