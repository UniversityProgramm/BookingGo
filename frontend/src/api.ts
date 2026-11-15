import type {
  Booking,
  ChangeEmail,
  ChangePassword,
  CreateBooking,
  CreateUser,
  NotificationsResponse,
  NotificationSettings,
  Role,
  UpdateProfile,
  User,
} from "./types";

const TOKEN_KEY = "bookinggo.token";
export const SESSION_EXPIRED = "bookinggo:session-expired";
export const tokenStore = {
  get: () => sessionStorage.getItem(TOKEN_KEY),
  set: (token: string) => sessionStorage.setItem(TOKEN_KEY, token),
  clear: () => sessionStorage.removeItem(TOKEN_KEY),
};

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

export async function request<T>(
  path: string,
  options: RequestInit = {},
  authenticated = true,
): Promise<T> {
  const token = authenticated ? tokenStore.get() : null;
  const headers = new Headers(options.headers);
  headers.set("Accept", "application/json");
  if (options.body) headers.set("Content-Type", "application/json");
  if (token) headers.set("Authorization", `Bearer ${token}`);
  let response: Response;
  try {
    response = await fetch(`/api${path}`, { ...options, headers });
  } catch (error) {
    if (error instanceof DOMException && error.name === "AbortError")
      throw error;
    throw new ApiError(
      0,
      "Не удалось связаться с сервером. Проверьте соединение и повторите попытку.",
    );
  }
  // An old in-flight request must never expire a newly created session.
  if (
    response.status === 401 &&
    authenticated &&
    token &&
    tokenStore.get() === token
  ) {
    tokenStore.clear();
    window.dispatchEvent(new Event(SESSION_EXPIRED));
  }
  const raw = await response.text();
  let data: unknown;
  try {
    data = raw ? JSON.parse(raw) : undefined;
  } catch {
    data = undefined;
  }
  if (!response.ok) {
    const message =
      data &&
      typeof data === "object" &&
      "error" in data &&
      typeof data.error === "string"
        ? data.error
        : null;
    throw new ApiError(
      response.status,
      response.status === 403
        ? "У вас нет доступа к этому действию."
        : message ||
          (response.status === 401
            ? "Сессия завершена. Войдите заново."
            : "Не удалось выполнить запрос. Повторите попытку."),
    );
  }
  if (raw && data === undefined)
    throw new ApiError(
      response.status,
      "Сервер вернул неожиданный ответ. Повторите попытку.",
    );
  return data as T;
}

const json = (method: string, body?: unknown): RequestInit => ({
  method,
  ...(body === undefined ? {} : { body: JSON.stringify(body) }),
});
export const managementPath = (role: Role) =>
  role === "admin" ? "/adminPanel" : "/staffPanel";
export const api = {
  login: (data: Pick<CreateUser, "email" | "password">) =>
    request<{ token: string }>("/auth/login", json("POST", data), false),
  register: (data: CreateUser) =>
    request<{ token: string }>("/auth/register", json("POST", data), false),
  logout: () => request("/auth/logout", json("POST")),
  me: (signal?: AbortSignal) => request<User>("/me/profile", { signal }),
  updateMe: (data: UpdateProfile) =>
    request<User>("/me/profile", json("PUT", data)),
  changeEmail: (data: ChangeEmail) =>
    request("/me/profile/email", json("PUT", data)),
  changePassword: (data: ChangePassword) =>
    request("/me/profile/password", json("PUT", data)),
  setupTotp: () =>
    request<{ manual_entry_key: string }>("/me/otp/setup", json("POST")),
  verifyTotp: (otp_code: string) =>
    request("/me/otp/verify", json("POST", { otp_code })),
  disableTotp: (otp_code: string, confirm_password: string) =>
    request("/me/otp/disable", json("POST", { otp_code, confirm_password })),
  bookings: (role: Role, signal?: AbortSignal) =>
    request<Booking[] | null>(
      role === "client" ? "/me/bookings" : `${managementPath(role)}/bookings`,
      { signal },
    ).then((data) => data ?? []),
  createBooking: (data: CreateBooking) =>
    request<Booking>("/me/bookings", json("POST", data)),
  cancelBooking: (id: number) =>
    request(`/me/bookings/cancelBooking/${id}`, json("POST")),
  completeBooking: (id: number) =>
    request<Booking>(
      `/staffPanel/bookings/completeBooking/${id}`,
      json("POST"),
    ),
  users: (role: Role, signal?: AbortSignal) =>
    request<User[] | null>(`${managementPath(role)}/users`, { signal }).then(
      (data) => data ?? [],
    ),
  createUser: (data: CreateUser) =>
    request<User>("/adminPanel/users", json("POST", data)),
  updateUser: (id: number, data: UpdateProfile) =>
    request<User>(`/adminPanel/users/${id}`, json("PUT", data)),
  deleteUser: (id: number) =>
    request(`/adminPanel/users/${id}`, json("DELETE")),
  notifications: (signal?: AbortSignal) =>
    request<NotificationsResponse>("/me/notifications", { signal }).then(
      normalizeNotifications,
    ),
  notificationSettings: (data: NotificationSettings) =>
    request("/me/notifications/settings", json("PATCH", data)),
};
export function normalizeNotifications(data: NotificationsResponse) {
  return Array.isArray(data) || data === null
    ? { notifications: data ?? [], warning: undefined }
    : { notifications: data.notifications ?? [], warning: data.warning };
}
