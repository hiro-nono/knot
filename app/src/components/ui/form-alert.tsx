export interface FormAlertProps {
  variant: "error" | "success";
  children: React.ReactNode;
}

export function FormAlert({ variant, children }: FormAlertProps) {
  const styles =
    variant === "error"
      ? "border-red-200 bg-red-50 text-red-700"
      : "border-primary/20 bg-primary-soft text-primary-hover";

  return (
    <div role={variant === "error" ? "alert" : "status"} className={`rounded-md border px-3 py-2 text-sm ${styles}`}>
      {children}
    </div>
  );
}
