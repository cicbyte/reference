/**
 * Path helpers.
 *
 * The Go backend returns OS-native paths (backslashes on Windows), but the UI
 * should display a consistent forward-slash form everywhere. We normalise for
 * *display* only — values passed back to the backend stay untouched, since Go
 * handles both separators correctly.
 */

/** Normalise a path to forward slashes for display. */
export function formatPath(p: string | undefined | null): string {
  if (!p) return ''
  return p.replace(/\\/g, '/')
}

/** Join path segments with a single forward slash (display only). */
export function joinPath(...parts: (string | undefined | null)[]): string {
  return parts
    .filter((p) => p != null && p !== '')
    .map((p, i) => {
      let s = String(p)
      s = s.replace(/\\/g, '/')
      if (i > 0) s = s.replace(/^\/+/, '')
      if (i < parts.length - 1) s = s.replace(/\/+$/, '')
      return s
    })
    .join('/')
}

/** Split a (possibly backslash) path into segments using '/'. */
export function pathSegments(p: string | undefined | null): string[] {
  return formatPath(p).split('/').filter(Boolean)
}
