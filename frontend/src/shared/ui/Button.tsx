import type { ButtonHTMLAttributes, PropsWithChildren } from "react";
import "./shared-ui.css";

type ButtonVariant = "primary" | "secondary" | "ghost";

interface ButtonProps
  extends PropsWithChildren, ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  fullWidth?: boolean;
}

export function Button({
  children,
  variant = "primary",
  fullWidth = false,
  className = "",
  ...props
}: ButtonProps) {
  return (
    <button
      className={`button button--${variant} ${fullWidth ? "button--full" : ""} ${className}`}
      {...props}
    >
      {children}
    </button>
  );
}
