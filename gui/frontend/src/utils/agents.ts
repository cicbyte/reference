/**
 * AI agent display-name registry. Single source of truth — replaces the
 * duplicated `{ claude: 'Claude', ... }` maps that were copy-pasted across
 * DashboardView / GlobalListView / GlobalStatsView / ProjectForm.
 */
export const agentNameMap: Record<string, string> = {
  claude: 'Claude',
  codex: 'Codex',
  zcode: 'ZCode',
  mimocode: 'MiMo',
  opencode: 'OpenCode',
}

/** Resolve an agent id to its display name, falling back to the id itself. */
export function agentDisplayName(id: string): string {
  return agentNameMap[id] || id
}
