export type Role = "client" | "staff" | "admin";
export type BookingStatus = "confirmed" | "cancelled" | "completed";
export interface NotificationSettings {
  is_email: boolean;
  is_sms: boolean;
}
export interface User {
  id: number;
  email: string;
  fio: string;
  phone: string;
  role: Role;
  created_at: string;
  updated_at: string;
  is_active: boolean;
  is_totp_enabled: boolean;
  notification_settings: NotificationSettings;
}
export interface Booking {
  id: number;
  user_id: number;
  slot_start: string;
  slot_end: string;
  status: BookingStatus;
  problem_description: string;
  created_at: string;
  updated_at: string;
}
export interface Notification {
  id: number;
  title: string;
  body: string;
  is_read: boolean;
  created_at: string;
}
export type NotificationsResponse =
  | Notification[]
  | null
  | { notifications: Notification[] | null; warning?: string };
export interface CreateUser {
  email: string;
  password: string;
  fio: string;
  phone: string;
}
export interface UpdateProfile {
  fio: string;
  phone: string;
}
export interface CreateBooking {
  slot_start: string;
  problem_description: string;
}
export interface ChangeEmail {
  new_email: string;
  confirm_password: string;
  otp_code?: string;
}
export interface ChangePassword {
  cur_password: string;
  new_password: string;
  otp_code?: string;
}
