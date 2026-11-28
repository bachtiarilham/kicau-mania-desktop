export type UserRole = "juri" | "operator" | "admin";

const AUTH_STORAGE_KEY = "kicau_desktop_auth_v1";

class AuthStore {
  isLoggedIn = $state(false);
  role = $state<UserRole | null>(null);
  userEmail = $state("");
  userName = $state("");
  userId = $state("");
  arenaId = $state("");

  constructor() {
    this.loadFromStorage();
  }

  private loadFromStorage() {
    if (typeof localStorage === "undefined") return;
    try {
      const raw = localStorage.getItem(AUTH_STORAGE_KEY);
      if (raw) {
        const parsed = JSON.parse(raw);
        if (parsed.isLoggedIn && parsed.role) {
          this.isLoggedIn = true;
          this.role = parsed.role;
          this.userEmail = parsed.userEmail || "";
          this.userName = parsed.userName || "";
          this.userId = parsed.userId || "";
          this.arenaId = parsed.arenaId || "";
        }
      }
    } catch {
      // Ignore parse error
    }
  }

  private saveToStorage() {
    if (typeof localStorage === "undefined") return;
    try {
      localStorage.setItem(
        AUTH_STORAGE_KEY,
        JSON.stringify({
          isLoggedIn: this.isLoggedIn,
          role: this.role,
          userEmail: this.userEmail,
          userName: this.userName,
          userId: this.userId,
          arenaId: this.arenaId,
        })
      );
    } catch {
      // Ignore write error
    }
  }

  setUserId(id: string) {
    this.userId = id;
    this.saveToStorage();
  }

  setArenaId(id: string) {
    this.arenaId = id;
    this.saveToStorage();
  }

  login(role: UserRole, email?: string, name?: string, userId?: string, arenaId?: string) {
    this.isLoggedIn = true;
    this.role = role;
    this.userEmail = email || (role === "juri" ? "juri@kicaumania.com" : role === "operator" ? "operator@kicaumania.com" : "admin@kicaumania.com");
    this.userName = name || (role === "juri" ? "Dewan Juri Arena" : role === "operator" ? "Operator Arena" : "Administrator");
    this.userId = userId || "";
    this.arenaId = arenaId || "";

    this.saveToStorage();
  }

  logout() {
    this.isLoggedIn = false;
    this.role = null;
    this.userEmail = "";
    this.userName = "";
    this.userId = "";
    this.arenaId = "";
    if (typeof localStorage !== "undefined") {
      localStorage.removeItem(AUTH_STORAGE_KEY);
    }
  }
}

export const auth = new AuthStore();
