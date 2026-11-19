import { describe, it, expect, beforeEach } from "vitest";
import { auth } from "./auth.svelte";

describe("auth store", () => {
  beforeEach(() => {
    auth.logout();
  });

  it("initializes as logged out", () => {
    expect(auth.isLoggedIn).toBe(false);
    expect(auth.role).toBeNull();
  });

  it("logs in as juri", () => {
    auth.login("juri");
    expect(auth.isLoggedIn).toBe(true);
    expect(auth.role).toBe("juri");
  });

  it("logs in as operator", () => {
    auth.login("operator");
    expect(auth.isLoggedIn).toBe(true);
    expect(auth.role).toBe("operator");
  });

  it("logs in as admin", () => {
    auth.login("admin");
    expect(auth.isLoggedIn).toBe(true);
    expect(auth.role).toBe("admin");
  });

  it("logs out cleanly", () => {
    auth.login("admin");
    expect(auth.isLoggedIn).toBe(true);
    auth.logout();
    expect(auth.isLoggedIn).toBe(false);
    expect(auth.role).toBeNull();
  });
});
