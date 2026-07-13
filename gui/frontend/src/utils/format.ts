/**
 * Formatting helpers. Replaces duplicated `fmtSize` / `formatSize` functions
 * that were copy-pasted across CacheReposView / GlobalStatsView / StorageForm.
 */

/**
 * Format a byte count as a human-readable size string.
 * - null/undefined → '…' (loading)
 * - 0              → '0 B'
 * - negative       → '—' (invalid)
 */
export function fmtSize(bytes: number | null | undefined | ''): string {
  if (bytes == null || bytes === '') return '…'
  if (bytes < 0) return '—'
  if (bytes === 0) return '0 B'
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  if (bytes < 1024 * 1024 * 1024) return (bytes / 1024 / 1024).toFixed(1) + ' MB'
  return (bytes / 1024 / 1024 / 1024).toFixed(2) + ' GB'
}
