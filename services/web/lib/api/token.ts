/**
 * Token storage helpers.
 *
 * We store the viewer JWT in localStorage so it survives page refreshes.
 * The key is intentionally non-generic to avoid collision with other apps
 * on the same origin.
 */

const TOKEN_KEY = "inori_auth_token";
const USER_KEY = "inori_auth_user";

export function storeToken(token: string) {
  if (typeof window !== "undefined") {
    localStorage.setItem(TOKEN_KEY, token);
  }
}

export function getStoredToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(TOKEN_KEY);
}

export function removeToken() {
  if (typeof window !== "undefined") {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
  }
}
