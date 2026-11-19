export type UserRole = "juri" | "operator" | "admin";

const AUTH_STORAGE_KEY = "kicau_desktop_auth_v1";

class AuthStore {
  isLoggedIn = $state(false);
  role = $state<UserRole | null>(null);
  userEmail = $state("");
  userName = $state("");

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
        })
      );
    } catch {
      // Ignore write error
    }
  }

  login(role: UserRole, email?: string, name?: string) {
    this.isLoggedIn = true;
    this.role = role;
    this.userEmail = email || (role === "juri" ? "juri@kicaumania.com" : role === "operator" ? "operator@kicaumania.com" : "admin@kicaumania.com");
    this.userName = name || (role === "juri" ? "Dewan Juri Arena" : role === "operator" ? "Operator Arena" : "Administrator");
    this.saveToStorage();
  }

  logout() {
    this.isLoggedIn = false;
    this.role = null;
    this.userEmail = "";
    this.userName = "";
    if (typeof localStorage !== "undefined") {
      localStorage.removeItem(AUTH_STORAGE_KEY);
    }
  }
}

export const auth = new AuthStore();
