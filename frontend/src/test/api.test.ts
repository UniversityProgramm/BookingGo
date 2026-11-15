import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  api,
  ApiError,
  normalizeNotifications,
  request,
  SESSION_EXPIRED,
  tokenStore,
} from "../api";
import { response } from "./fixtures";

describe("HTTP client", () => {
  beforeEach(() => tokenStore.set("test-session"));
  it("uses /api and sends a Bearer token and JSON headers", async () => {
    const fetch = vi.fn().mockResolvedValue(response({ ok: true }));
    vi.stubGlobal("fetch", fetch);
    await request("/me/profile", {
      method: "PUT",
      body: JSON.stringify({ fio: "Иван" }),
    });
    const [url, options] = fetch.mock.calls[0];
    expect(url).toBe("/api/me/profile");
    expect(options.headers.get("Authorization")).toBe("Bearer test-session");
    expect(options.headers.get("Content-Type")).toBe("application/json");
  });
  it("does not attach an existing token to public login requests", async () => {
    const fetch = vi.fn().mockResolvedValue(response({ token: "new-session" }));
    vi.stubGlobal("fetch", fetch);
    await api.login({ email: "test@example.com", password: "test-password" });
    expect(fetch.mock.calls[0][1].headers.has("Authorization")).toBe(false);
  });
  it("accepts an empty 204 response", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(response(undefined, 204)));
    await expect(api.cancelBooking(1)).resolves.toBeUndefined();
  });
  it("clears expired sessions and emits one expiration event", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(response({ error: "expired" }, 401)),
    );
    const listener = vi.fn();
    window.addEventListener(SESSION_EXPIRED, listener);
    await expect(api.me()).rejects.toMatchObject({ status: 401 });
    expect(tokenStore.get()).toBeNull();
    expect(listener).toHaveBeenCalledTimes(1);
    window.removeEventListener(SESSION_EXPIRED, listener);
  });
  it("keeps the session on 403", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(response({ error: "forbidden" }, 403)),
    );
    await expect(api.me()).rejects.toThrow("У вас нет доступа");
    expect(tokenStore.get()).toBe("test-session");
  });
  it("does not expire a session for an invalid login password", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(response({ error: "Неверный пароль" }, 401)),
    );
    await expect(
      api.login({ email: "test@example.com", password: "incorrect" }),
    ).rejects.toThrow("Неверный пароль");
    expect(tokenStore.get()).toBe("test-session");
  });
  it("ignores a delayed 401 from the previous session", async () => {
    let resolve!: (response: Response) => void;
    vi.stubGlobal(
      "fetch",
      vi.fn().mockImplementation(
        () =>
          new Promise<Response>((done) => {
            resolve = done;
          }),
      ),
    );
    const pending = api.me();
    tokenStore.set("new-session");
    resolve(response({ error: "expired" }, 401));
    await expect(pending).rejects.toMatchObject({ status: 401 });
    expect(tokenStore.get()).toBe("new-session");
  });
  it("reports network failures without clearing the session", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockRejectedValue(new TypeError("Failed to fetch")),
    );
    await expect(api.me()).rejects.toMatchObject({ status: 0 });
    expect(tokenStore.get()).toBe("test-session");
  });
  it("keeps aborts as aborts", async () => {
    const error = new DOMException("Aborted", "AbortError");
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(error));
    await expect(api.me()).rejects.toBe(error);
  });
  it("reports a malformed success response", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(new Response("<html>proxy error</html>")),
    );
    await expect(api.me()).rejects.toBeInstanceOf(ApiError);
  });
  it("keeps server conflict details", async () => {
    vi.stubGlobal(
      "fetch",
      vi
        .fn()
        .mockResolvedValue(
          response({ error: "Эти дата и время недоступны для записи" }, 400),
        ),
    );
    await expect(
      api.createBooking({
        slot_start: "2099-01-01T10:00:00Z",
        problem_description: "",
      }),
    ).rejects.toThrow("недоступны");
  });
});

describe("notification response normalization", () => {
  const notification = {
    id: 1,
    title: "Запись создана",
    body: "Готово",
    is_read: false,
    created_at: "2026-01-01T12:00:00Z",
  };
  it("supports the plain array response", () =>
    expect(normalizeNotifications([notification]).notifications).toEqual([
      notification,
    ]));
  it("supports the envelope and preserves its warning", () =>
    expect(
      normalizeNotifications({
        notifications: [notification],
        warning: "Уже прочитаны",
      }),
    ).toEqual({ notifications: [notification], warning: "Уже прочитаны" }));
  it("supports Go nil slices", () => {
    expect(normalizeNotifications(null).notifications).toEqual([]);
    expect(
      normalizeNotifications({ notifications: null }).notifications,
    ).toEqual([]);
  });
});
