import { toast } from "sonner";

export function handleError(error: unknown, context: string): void {
  const message = error instanceof Error ? error.message : String(error);
  console.error(`${context}:`, error);
  toast.error(`${context}: ${message}`);
}
