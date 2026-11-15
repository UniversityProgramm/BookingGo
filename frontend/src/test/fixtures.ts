import type { Booking, Role, User } from "../types";

export const makeUser = (role: Role = "client"): User => ({
  id: 1,
  email: "test@example.com",
  fio: "Иван Петров",
  phone: "+79990000000",
  role,
  created_at: "2026-01-01T12:00:00Z",
  updated_at: "2026-01-01T12:00:00Z",
  is_active: true,
  is_totp_enabled: false,
  notification_settings: { is_email: true, is_sms: false },
});
export const makeBooking = (overrides: Partial<Booking> = {}): Booking => ({
  id: 1,
  user_id: 1,
  slot_start: "2099-01-15T12:00:00Z",
  slot_end: "2099-01-15T13:00:00Z",
  problem_description: "Консультация",
  status: "confirmed",
  created_at: "2026-01-01T12:00:00Z",
  updated_at: "2026-01-01T12:00:00Z",
  ...overrides,
});
export function response(data: unknown, status = 200) {
  return new Response(data === undefined ? null : JSON.stringify(data), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}
