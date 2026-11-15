import {
  render,
  screen,
  waitFor,
  within,
  fireEvent,
} from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";
import { App } from "../App";
import { AuthProvider } from "../auth";
import { tokenStore } from "../api";
import type { Role } from "../types";
import { makeBooking, makeUser, response } from "./fixtures";

vi.mock("qrcode", () => ({
  default: {
    toDataURL: vi.fn().mockResolvedValue("data:image/png;base64,dGVzdA=="),
  },
}));

function mount(
  path: string,
  {
    role = "client",
    authenticated = true,
    handler = () => response([]),
  }: {
    role?: Role;
    authenticated?: boolean;
    handler?: (
      url: string,
      options: RequestInit,
    ) => Response | Promise<Response>;
  } = {},
) {
  if (authenticated) tokenStore.set("test-session");
  const userProfile = makeUser(role);
  const fetch = vi
    .fn()
    .mockImplementation((url: string, options: RequestInit) =>
      url === "/api/me/profile" && (!options.method || options.method === "GET")
        ? Promise.resolve(response(userProfile))
        : Promise.resolve(handler(url, options)),
    );
  vi.stubGlobal("fetch", fetch);
  const client = new QueryClient({
    defaultOptions: {
      queries: { retry: false, gcTime: 0 },
      mutations: { retry: false },
    },
  });
  render(
    <QueryClientProvider client={client}>
      <MemoryRouter initialEntries={[path]}>
        <AuthProvider>
          <App />
        </AuthProvider>
      </MemoryRouter>
    </QueryClientProvider>,
  );
  return { client, fetch, user: userEvent.setup(), userProfile };
}
describe("authentication and role routes", () => {
  it("redirects anonymous deep links to login", async () => {
    mount("/admin/users", { authenticated: false });
    expect(
      await screen.findByRole("heading", { name: "С возвращением" }),
    ).toBeInTheDocument();
  });
  it.each<Role>(["client", "staff", "admin"])(
    "restores %s profile and opens its home",
    async (role) => {
      const { fetch } = mount("/", { role });
      expect(
        await screen.findByRole("heading", {
          name: role === "client" ? "Мои записи" : "Все записи",
        }),
      ).toBeInTheDocument();
      await waitFor(() =>
        expect(fetch).toHaveBeenCalledWith(
          role === "client" ? "/api/me/bookings" : `/api/${role}Panel/bookings`,
          expect.anything(),
        ),
      );
    },
  );
  it.each<[Role, string]>([
    ["client", "/admin/users"],
    ["client", "/staff/bookings"],
    ["staff", "/admin/users"],
    ["admin", "/staff/bookings"],
    ["staff", "/client/bookings/new"],
  ])("denies %s on %s", async (role, path) => {
    const { fetch } = mount(path, { role });
    expect(
      await screen.findByRole("heading", { name: "Нет доступа" }),
    ).toBeInTheDocument();
    expect(fetch).toHaveBeenCalledTimes(1);
    expect(tokenStore.get()).toBe("test-session");
  });
  it("clears private query data and redirects when a request expires", async () => {
    const { client } = mount("/client/bookings", {
      handler: () => response({ error: "expired" }, 401),
    });
    client.setQueryData(["private-data"], { private: true });
    expect(
      await screen.findByRole("heading", { name: "С возвращением" }),
    ).toBeInTheDocument();
    expect(client.getQueryData(["private-data"])).toBeUndefined();
    expect(tokenStore.get()).toBeNull();
    expect(
      screen.getByText("Сессия завершена. Войдите заново."),
    ).toBeInTheDocument();
  });
  it("shows 403 without signing out", async () => {
    mount("/client/bookings", {
      handler: () => response({ error: "forbidden" }, 403),
    });
    expect(
      await screen.findByText("У вас нет доступа к этому действию."),
    ).toBeInTheDocument();
    expect(tokenStore.get()).toBe("test-session");
  });
  it("does not fetch read-marking notifications outside their section", async () => {
    const { fetch, user } = mount("/profile");
    await screen.findByRole("heading", { name: "Мой профиль" });
    expect(
      fetch.mock.calls.some(([url]) => url === "/api/me/notifications"),
    ).toBe(false);
    await user.click(screen.getByRole("link", { name: "Уведомления" }));
    expect(
      await screen.findByRole("heading", { name: "Вы всё прочитали" }),
    ).toBeInTheDocument();
    expect(
      fetch.mock.calls.filter(([url]) => url === "/api/me/notifications"),
    ).toHaveLength(1);
  });
});
describe("booking workflows", () => {
  it("preserves all fields on a slot conflict and sends RFC3339", async () => {
    const { fetch, user } = mount("/client/bookings/new", {
      handler: () =>
        response({ error: "Эти дата и время недоступны для записи" }, 400),
    });
    await screen.findByLabelText("Дата");
    fireEvent.change(screen.getByLabelText("Дата"), {
      target: { value: "2099-01-15" },
    });
    fireEvent.change(screen.getByLabelText("Время начала"), {
      target: { value: "14:30" },
    });
    await user.type(
      screen.getByRole("textbox", { name: /Описание обращения/ }),
      "Мой вопрос",
    );
    await user.click(screen.getByRole("button", { name: "Создать запись" }));
    expect(await screen.findByText(/Это время уже занято/)).toBeInTheDocument();
    expect(screen.getByLabelText("Дата")).toHaveValue("2099-01-15");
    expect(screen.getByLabelText("Время начала")).toHaveValue("14:30");
    expect(
      screen.getByRole("textbox", { name: /Описание обращения/ }),
    ).toHaveValue("Мой вопрос");
    const payload = JSON.parse(
      fetch.mock.calls.find(([url]) => url === "/api/me/bookings")![1].body,
    );
    expect(payload.slot_start).toBe(new Date("2099-01-15T14:30").toISOString());
  });
  it("blocks duplicate submissions while creating a booking", async () => {
    const { fetch, user } = mount("/client/bookings/new", {
      handler: () => new Promise(() => {}),
    });
    await screen.findByLabelText("Дата");
    fireEvent.change(screen.getByLabelText("Дата"), {
      target: { value: "2099-01-15" },
    });
    fireEvent.change(screen.getByLabelText("Время начала"), {
      target: { value: "14:30" },
    });
    await user.dblClick(screen.getByRole("button", { name: "Создать запись" }));
    expect(screen.getByRole("button", { name: "Подождите…" })).toBeDisabled();
    expect(
      fetch.mock.calls.filter(([url]) => url === "/api/me/bookings"),
    ).toHaveLength(1);
  });
  it("requires confirmation before canceling a future booking", async () => {
    let canceled = false;
    const { fetch, user } = mount("/client/bookings", {
      handler: (url) => {
        if (url.includes("cancelBooking")) {
          canceled = true;
          return response(undefined, 204);
        }
        return response([
          makeBooking({ status: canceled ? "cancelled" : "confirmed" }),
        ]);
      },
    });
    await user.click(await screen.findByRole("button", { name: "Отменить" }));
    expect(
      fetch.mock.calls.some(([url]) => url.includes("cancelBooking")),
    ).toBe(false);
    await user.click(
      within(screen.getByRole("dialog")).getByRole("button", {
        name: "Отменить запись",
      }),
    );
    expect(await screen.findByText("Отменена")).toBeInTheDocument();
  });
  it("only shows completion for a staff member after the booking ends", async () => {
    const { user } = mount("/staff/bookings", {
      role: "staff",
      handler: (url) =>
        response(
          url.endsWith("/users")
            ? []
            : [
                makeBooking(),
                makeBooking({
                  id: 2,
                  slot_start: "2020-01-15T12:00:00Z",
                  slot_end: "2020-01-15T13:00:00Z",
                }),
              ],
        ),
    });
    expect(
      await screen.findAllByRole("button", { name: "Завершить" }),
    ).toHaveLength(1);
    await user.click(screen.getByRole("button", { name: "Отменённые" }));
    expect(
      await screen.findByRole("heading", { name: "Записи не найдены" }),
    ).toBeInTheDocument();
  });
});
describe("profile credential changes", () => {
  it.each(["email", "password"])(
    "ends the client session after changing %s",
    async (operation) => {
      const { client, user } = mount("/security", {
        handler: () => response({ message: "success" }),
      });
      await screen.findByRole("heading", { name: "Безопасность" });
      client.setQueryData(["private-data"], "private");
      const section = screen
        .getByRole("heading", {
          name: operation === "email" ? "Изменить почту" : "Изменить пароль",
        })
        .closest("section")!;
      await user.type(
        within(section).getByLabelText("Текущий пароль"),
        "current-password",
      );
      await user.type(
        within(section).getByLabelText(
          operation === "email" ? "Новая электронная почта" : /Новый пароль/,
        ),
        operation === "email" ? "new@example.com" : "new-password",
      );
      await user.click(
        within(section).getByRole("button", {
          name: operation === "email" ? "Изменить почту" : "Изменить пароль",
        }),
      );
      expect(
        await screen.findByRole("heading", { name: "С возвращением" }),
      ).toBeInTheDocument();
      expect(tokenStore.get()).toBeNull();
      expect(client.getQueryData(["private-data"])).toBeUndefined();
    },
  );
});

describe("management and TOTP workflows", () => {
  it("lets staff inspect users without edit or delete actions", async () => {
    const { user } = mount("/staff/users", {
      role: "staff",
      handler: () => response([makeUser()]),
    });
    await user.click(await screen.findByRole("button", { name: "Подробнее" }));
    expect(screen.getByRole("dialog")).toBeInTheDocument();
    expect(screen.getByLabelText("ФИО")).toHaveAttribute("readonly");
    expect(
      screen.queryByRole("button", { name: "Удалить" }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Сохранить изменения" }),
    ).not.toBeInTheDocument();
  });
  it("lets an administrator search, create, edit and confirm deletion", async () => {
    let users = [makeUser()];
    const { fetch, user } = mount("/admin/users", {
      role: "admin",
      handler: (url, options) => {
        if (options.method === "POST") {
          const payload = JSON.parse(options.body as string);
          users.push({ ...makeUser(), ...payload, id: 2 });
          return response(users[1], 201);
        }
        if (options.method === "PUT") {
          users[1] = { ...users[1], ...JSON.parse(options.body as string) };
          return response(users[1]);
        }
        if (options.method === "DELETE") {
          users = users.filter((item) => item.id !== 2);
          return response(undefined, 204);
        }
        return response(users);
      },
    });
    await user.click(
      await screen.findByRole("button", { name: "Создать клиента" }),
    );
    let dialog = within(screen.getByRole("dialog"));
    await user.type(dialog.getByLabelText("ФИО"), "Анна Смирнова");
    await user.type(dialog.getByLabelText("Телефон"), "+79991111111");
    await user.type(
      dialog.getByLabelText("Электронная почта"),
      "anna@example.com",
    );
    await user.type(dialog.getByLabelText("Пароль"), "test-password");
    await user.click(dialog.getByRole("button", { name: "Создать клиента" }));
    expect(await screen.findByText("Анна Смирнова")).toBeInTheDocument();
    await user.type(
      screen.getByLabelText("Поиск пользователей"),
      "anna@example.com",
    );
    expect(
      screen.queryByText("Иван Петров", { selector: "td strong" }),
    ).not.toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Изменить" }));
    dialog = within(screen.getByRole("dialog"));
    await user.clear(dialog.getByLabelText("ФИО"));
    await user.type(dialog.getByLabelText("ФИО"), "Анна Иванова");
    await user.click(
      dialog.getByRole("button", { name: "Сохранить изменения" }),
    );
    expect(await screen.findByText("Анна Иванова")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Удалить" }));
    expect(
      fetch.mock.calls.some(([, options]) => options.method === "DELETE"),
    ).toBe(false);
    await user.click(
      within(screen.getByRole("dialog")).getByRole("button", {
        name: "Удалить",
      }),
    );
    expect(
      await screen.findByRole("heading", { name: "Пользователи не найдены" }),
    ).toBeInTheDocument();
  });
  it("sets up a local QR code, verifies and disables TOTP", async () => {
    const { fetch, user, userProfile } = mount("/security", {
      handler: (url) => {
        if (url.endsWith("/setup"))
          return response({
            manual_entry_key:
              "otpauth://totp/BookingGo:test?secret=JBSWY3DPEHPK3PXP&issuer=BookingGo",
          });
        if (url.endsWith("/verify")) userProfile.is_totp_enabled = true;
        if (url.endsWith("/disable")) userProfile.is_totp_enabled = false;
        return response({ message: "success" });
      },
    });
    await user.click(
      await screen.findByRole("button", { name: "Настроить защиту" }),
    );
    const qr = await screen.findByRole("img", { name: /QR-код/ });
    await waitFor(() =>
      expect(qr).toHaveAttribute(
        "src",
        expect.stringMatching(/^data:image\/png/),
      ),
    );
    await user.type(
      screen.getByLabelText("Код из приложения-аутентификатора"),
      "123456",
    );
    await user.click(screen.getByRole("button", { name: "Включить защиту" }));
    expect(
      await screen.findByText("Двухфакторная защита включена."),
    ).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Отключить защиту" }));
    const panel = screen
      .getByRole("heading", { name: "Двухфакторная защита" })
      .closest("section")!;
    await user.type(
      within(panel).getByLabelText("Текущий пароль"),
      "test-password",
    );
    await user.type(
      within(panel).getByLabelText("Код из приложения-аутентификатора"),
      "123456",
    );
    await user.click(
      within(panel).getByRole("button", { name: "Отключить защиту" }),
    );
    expect(
      await screen.findByText("Двухфакторная защита отключена."),
    ).toBeInTheDocument();
    expect(fetch.mock.calls.every(([url]) => url.startsWith("/api/"))).toBe(
      true,
    );
  });
  it("renders the notifications envelope and its warning", async () => {
    mount("/notifications", {
      handler: () =>
        response({
          notifications: [
            {
              id: 1,
              title: "Новая запись",
              body: "Визит подтверждён",
              is_read: true,
              created_at: "2026-01-01T12:00:00Z",
            },
          ],
          warning: "Все уведомления уже прочитаны",
        }),
    });
    expect(
      await screen.findByText("Все уведомления уже прочитаны"),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("heading", { name: "Новая запись" }),
    ).toBeInTheDocument();
  });
  it("lets a failed network query be retried", async () => {
    let attempts = 0;
    const { user } = mount("/client/bookings", {
      handler: () => {
        attempts++;
        if (attempts === 1) throw new TypeError("Offline");
        return response([makeBooking()]);
      },
    });
    await user.click(
      await screen.findByRole("button", { name: "Повторить попытку" }),
    );
    expect(await screen.findByText("Консультация")).toBeInTheDocument();
    expect(attempts).toBe(2);
  });
});
