import { type SelectHTMLAttributes, forwardRef, useId } from "react";

export interface SelectFieldProps extends SelectHTMLAttributes<HTMLSelectElement> {
  label: string;
}

export const SelectField = forwardRef<HTMLSelectElement, SelectFieldProps>(function SelectField(
  { label, id, className = "", children, ...props },
  ref,
) {
  const generatedId = useId();
  const selectId = id ?? generatedId;

  return (
    <div className="flex flex-col gap-1.5">
      <label htmlFor={selectId} className="text-sm font-medium text-zinc-700">
        {label}
      </label>
      <select
        ref={ref}
        id={selectId}
        className={`rounded-md border border-zinc-300 bg-white px-3 py-2 text-sm text-zinc-900 outline-none focus:ring-2 focus:ring-primary ${className}`}
        {...props}
      >
        {children}
      </select>
    </div>
  );
});
