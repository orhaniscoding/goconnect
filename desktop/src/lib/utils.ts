import { toast } from "sonner";

export function handleError(error: unknown, context: string): void {
  const message = error instanceof Error ? error.message : String(error);
  console.error(`${context}:`, error);
  toast.error(`${context}: ${message}`);
}

/**
 * Format message text with simple Markdown support
 * Supports: **bold**, *italic*, `code`, @mentions
 * XSS-safe: escapes HTML before applying formatting
 */
export function formatMarkdown(text: string): string {
  // Escape HTML first to prevent XSS
  let safe = text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;');

  // Apply markdown formatting
  // Order matters: bold before italic to handle ***text*** correctly
  return safe
    // Bold: **text**
    .replace(/\*\*(.+?)\*\*/g, '<strong class="font-bold">$1</strong>')
    // Italic: *text*
    .replace(/\*(.+?)\*/g, '<em class="italic">$1</em>')
    // Inline code: `code`
    .replace(/`([^`]+)`/g, '<code class="inline-code">$1</code>')
    // Mentions: @username
    .replace(/@(\w+)/g, '<span class="mention">@$1</span>')
    // Newlines to <br>
    .replace(/\n/g, '<br>');
}

/**
 * Check if a message can be edited (within 15 minute window)
 */
export function canEditMessage(timestamp: string): boolean {
  const EDIT_WINDOW_MS = 15 * 60 * 1000; // 15 minutes
  const messageTime = new Date(timestamp).getTime();
  return Date.now() - messageTime < EDIT_WINDOW_MS;
}

/**
 * Format timestamp for display
 */
export function formatMessageTime(timestamp: string): string {
  const date = new Date(timestamp);
  const now = new Date();
  const isToday = date.toDateString() === now.toDateString();
  
  if (isToday) {
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  }
  
  const yesterday = new Date(now);
  yesterday.setDate(yesterday.getDate() - 1);
  const isYesterday = date.toDateString() === yesterday.toDateString();
  
  if (isYesterday) {
    return `Yesterday, ${date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}`;
  }
  
  return date.toLocaleDateString([], { month: 'short', day: 'numeric' }) + 
    ', ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}
