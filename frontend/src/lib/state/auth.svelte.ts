export type UserRole = "juri" | "operator" | "admin";

const AUTH_STORAGE_KEY = "kicau_desktop_auth_v1";

class AuthStore {
  isLoggedIn = $state(false);
  role = $state<UserRole | null>(null);
  userEmail = $state("");
  userName = $state("");
  arenaId = $state("");
  judgeIndex = $state(0); // 0 = Juri 1, 1 = Juri 2, 2 = Juri 3, 3 = Juri 4

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
          this.arenaId = parsed.arenaId || "";
          if (typeof parsed.judgeIndex === "number") {
            this.judgeIndex = parsed.judgeIndex;
          }
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
          arenaId: this.arenaId,
          judgeIndex: this.judgeIndex,
        })
      );
    } catch {
      // Ignore write error
    }
  }

  setJudgeIndex(idx: number) {
    this.judgeIndex = Math.max(0, Math.min(3, idx));
    this.saveToStorage();
  }

  setArenaId(id: string) {
    this.arenaId = id;
    this.saveToStorage();
  }

  login(role: UserRole, email?: string, name?: string, judgeIdx?: number, arenaId?: string) {
    this.isLoggedIn = true;
    this.role = role;
    this.userEmail = email || (role === "juri" ? "juri@kicaumania.com" : role === "operator" ? "operator@kicaumania.com" : "admin@kicaumania.com");
    this.userName = name || (role === "juri" ? "Dewan Juri Arena" : role === "operator" ? "Operator Arena" : "Administrator");
    this.arenaId = arenaId || "";

    if (typeof judgeIdx === "number") {
      this.judgeIndex = Math.max(0, Math.min(3, judgeIdx));
    } else if (role === "juri") {
      const combined = (this.userEmail + " " + this.userName).toLowerCase();
      if (combined.includes("juri 4") || combined.includes("juri4")) {
        this.judgeIndex = 3;
      } else if (combined.includes("juri 3") || combined.includes("juri3")) {
        this.judgeIndex = 2;
      } else if (combined.includes("juri 2") || combined.includes("juri2")) {
        this.judgeIndex = 1;
      } else {
        this.judgeIndex = 0;
      }
    }

    this.saveToStorage();
  }

  logout() {
    this.isLoggedIn = false;
    this.role = null;
    this.userEmail = "";
    this.userName = "";
    this.arenaId = "";
    if (typeof localStorage !== "undefined") {
      localStorage.removeItem(AUTH_STORAGE_KEY);
    }
  }
}

export const auth = new AuthStore();
