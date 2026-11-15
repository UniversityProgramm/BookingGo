import { useState, type FormEvent } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Bell, Mail, Smartphone, UserRound } from "lucide-react";
import { api } from "../api";
import { roleLabels, useAuth } from "../auth";
import { Feedback, formatDate, PageHeading, Submit } from "../components/ui";

export function ProfilePage() {
  const { user } = useAuth();
  const cache = useQueryClient();
  const [form, setForm] = useState({ fio: user!.fio, phone: user!.phone });
  const [settings, setSettings] = useState(
    user!.notification_settings ?? { is_email: false, is_sms: false },
  );
  const profile = useMutation({
    mutationFn: () => api.updateMe(form),
    onSuccess: () => cache.invalidateQueries({ queryKey: ["profile"] }),
  });
  const notifications = useMutation({
    mutationFn: () => api.notificationSettings(settings),
    onSuccess: () => cache.invalidateQueries({ queryKey: ["profile"] }),
  });
  const saveProfile = (event: FormEvent) => {
    event.preventDefault();
    if (!profile.isPending) profile.mutate();
  };
  const saveNotifications = (event: FormEvent) => {
    event.preventDefault();
    if (!notifications.isPending) notifications.mutate();
  };
  return (
    <>
      <PageHeading
        eyebrow="ЛИЧНЫЕ ДАННЫЕ"
        title="Мой профиль"
        description="Актуальные контакты помогают оставаться на связи."
      />
      <div className="profile-grid">
        <section className="panel form-panel">
          <div className="section-heading">
            <UserRound size={21} />
            <h2>Контактные данные</h2>
          </div>
          <form className="stack" onSubmit={saveProfile}>
            <fieldset disabled={profile.isPending}>
              <label>
                ФИО
                <input
                  autoComplete="name"
                  value={form.fio}
                  onChange={(event) =>
                    setForm({ ...form, fio: event.target.value })
                  }
                  required
                />
              </label>
              <label>
                Телефон
                <input
                  autoComplete="tel"
                  type="tel"
                  value={form.phone}
                  onChange={(event) =>
                    setForm({ ...form, phone: event.target.value })
                  }
                  required
                  minLength={10}
                  maxLength={15}
                />
              </label>
              <label>
                Электронная почта
                <input type="email" value={user!.email} readOnly />
                <small className="muted">
                  Изменить почту можно в разделе «Безопасность».
                </small>
              </label>
            </fieldset>
            <Feedback
              error={profile.error}
              success={profile.isSuccess ? "Профиль сохранён." : ""}
            />
            <div>
              <Submit pending={profile.isPending}>Сохранить изменения</Submit>
            </div>
          </form>
        </section>
        <aside className="panel profile-card">
          <span className="avatar large">
            {user!.fio.trim().charAt(0) || "Б"}
          </span>
          <h2>{user!.fio}</h2>
          <p className="muted">{user!.email}</p>
          <span className="role-pill">{roleLabels[user!.role]}</span>
          <div className="profile-meta">
            <span>В BookingGo с</span>
            <strong>{formatDate(user!.created_at)}</strong>
          </div>
        </aside>
      </div>
      <section className="panel form-panel notification-settings">
        <div className="section-heading">
          <Bell size={21} />
          <h2>Настройки уведомлений</h2>
        </div>
        <p className="muted">Выберите каналы для сообщений о ваших записях.</p>
        <form onSubmit={saveNotifications} className="stack">
          <fieldset disabled={notifications.isPending}>
            <label className="setting-row">
              <Mail size={22} />
              <span>
                <strong>Электронная почта</strong>
                <small>Сообщения на {user!.email}</small>
              </span>
              <input
                type="checkbox"
                checked={settings.is_email}
                onChange={(event) =>
                  setSettings({ ...settings, is_email: event.target.checked })
                }
              />
            </label>
            <label className="setting-row">
              <Smartphone size={22} />
              <span>
                <strong>SMS</strong>
                <small>Сообщения на {user!.phone}</small>
              </span>
              <input
                type="checkbox"
                checked={settings.is_sms}
                onChange={(event) =>
                  setSettings({ ...settings, is_sms: event.target.checked })
                }
              />
            </label>
          </fieldset>
          <p className="small muted">
            Доставка email и SMS зависит от настройки сервиса. Уведомления в
            личном кабинете доступны всегда.
          </p>
          <Feedback
            error={notifications.error}
            success={
              notifications.isSuccess ? "Настройки уведомлений сохранены." : ""
            }
          />
          <div>
            <Submit pending={notifications.isPending}>
              Сохранить настройки
            </Submit>
          </div>
        </form>
      </section>
    </>
  );
}
