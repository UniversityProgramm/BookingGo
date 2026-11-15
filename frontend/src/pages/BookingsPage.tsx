import { useEffect, useState, type FormEvent } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  ArrowLeft,
  ArrowRight,
  CalendarDays,
  CheckCheck,
  Clock3,
  Plus,
  Search,
} from "lucide-react";
import { api, ApiError } from "../api";
import { roleHome, useAuth } from "../auth";
import type { Booking, BookingStatus } from "../types";
import {
  Confirm,
  Empty,
  ErrorState,
  Feedback,
  formatDate,
  formatDateTime,
  formatTime,
  Loading,
  PageHeading,
  Submit,
  TimezoneNote,
} from "../components/ui";

export const statusLabels: Record<BookingStatus, string> = {
  confirmed: "Подтверждена",
  cancelled: "Отменена",
  completed: "Завершена",
};
export function BookingsPage() {
  const { user } = useAuth();
  const cache = useQueryClient();
  const [status, setStatus] = useState("all");
  const [search, setSearch] = useState("");
  const [confirm, setConfirm] = useState<Booking | null>(null);
  const [now, setNow] = useState(Date.now());
  useEffect(() => {
    const timer = window.setInterval(() => setNow(Date.now()), 30_000);
    return () => window.clearInterval(timer);
  }, []);
  const role = user!.role;
  const query = useQuery({
    queryKey: ["bookings", role, user!.id],
    queryFn: ({ signal }) => api.bookings(role, signal),
  });
  const users = useQuery({
    queryKey: ["users", role],
    queryFn: ({ signal }) => api.users(role, signal),
    enabled: role !== "client",
  });
  const bookings = query.data ?? [];
  const visible = bookings
    .filter(
      (item) =>
        (status === "all" || status === item.status) &&
        `${item.id} ${item.user_id} ${item.problem_description} ${users.data?.find((user) => user.id === item.user_id)?.fio ?? ""}`
          .toLocaleLowerCase("ru")
          .includes(search.toLocaleLowerCase("ru")),
    )
    .sort(
      (a, b) =>
        new Date(b.slot_start).getTime() - new Date(a.slot_start).getTime(),
    );
  const upcoming = bookings.filter(
    (item) =>
      item.status === "confirmed" && new Date(item.slot_start).getTime() > now,
  );
  const next = [...upcoming].sort((a, b) =>
    a.slot_start.localeCompare(b.slot_start),
  )[0];
  return (
    <>
      <PageHeading
        eyebrow={
          role === "client" ? "ВАШЕ РАСПИСАНИЕ" : "УПРАВЛЕНИЕ РАСПИСАНИЕМ"
        }
        title={role === "client" ? "Мои записи" : "Все записи"}
        description={
          role === "client"
            ? "Планируйте визиты и управляйте своим временем."
            : "Следите за расписанием и статусами записей."
        }
        action={
          role === "client" && (
            <Link className="button" to="/client/bookings/new">
              <Plus size={18} />
              Новая запись
            </Link>
          )
        }
      />
      <section className="stats-grid" aria-label="Сводка записей">
        <div className="stat-card">
          <span className="stat-icon blue">
            <CalendarDays size={21} />
          </span>
          <div>
            <p>Предстоящие</p>
            <strong>{query.isPending ? "—" : upcoming.length}</strong>
            <small>записей в расписании</small>
          </div>
        </div>
        <div className="stat-card">
          <span className="stat-icon green">
            <CheckCheck size={21} />
          </span>
          <div>
            <p>Завершённые</p>
            <strong>
              {query.isPending
                ? "—"
                : bookings.filter((item) => item.status === "completed").length}
            </strong>
            <small>успешных визитов</small>
          </div>
        </div>
        <div className="stat-card next-stat">
          <span className="stat-icon violet">
            <Clock3 size={21} />
          </span>
          <div>
            <p>Ближайшая запись</p>
            <strong>{next ? formatTime(next.slot_start) : "—"}</strong>
            <small>
              {next ? formatDate(next.slot_start) : "Новых записей пока нет"}
            </small>
          </div>
        </div>
      </section>
      <section className="panel">
        <div className="panel-title">
          <div>
            <h2>История записей</h2>
            <p className="muted">
              {role === "client"
                ? "Все ваши визиты и их текущий статус"
                : "Записи всех клиентов"}
            </p>
          </div>
          <span className="count">{bookings.length}</span>
        </div>
        <div className="toolbar">
          <div className="tabs" aria-label="Фильтр по статусу">
            {[
              ["all", "Все записи"],
              ["confirmed", "Подтверждённые"],
              ["completed", "Завершённые"],
              ["cancelled", "Отменённые"],
            ].map(([value, label]) => (
              <button
                key={value}
                className={status === value ? "active" : ""}
                aria-pressed={status === value}
                onClick={() => setStatus(value)}
              >
                {label}
              </button>
            ))}
          </div>
          {role !== "client" && (
            <label className="search">
              <Search size={17} />
              <input
                aria-label="Поиск записей"
                placeholder="Номер, клиент или описание"
                value={search}
                onChange={(event) => setSearch(event.target.value)}
              />
            </label>
          )}
        </div>
        {query.isPending ? (
          <Loading />
        ) : query.error ? (
          <ErrorState
            error={query.error}
            retry={() => {
              void query.refetch();
            }}
          />
        ) : visible.length === 0 ? (
          <Empty
            title={
              bookings.length ? "Записи не найдены" : "Начните с первой записи"
            }
          >
            {bookings.length
              ? "Попробуйте изменить поиск или выбранный статус."
              : role === "client"
                ? "Выберите удобную дату и время — визит появится здесь."
                : "Новые записи появятся в этом разделе."}
            {!bookings.length && role === "client" && (
              <Link
                className="button secondary empty-action"
                to="/client/bookings/new"
              >
                Создать запись
                <ArrowRight size={16} />
              </Link>
            )}
          </Empty>
        ) : (
          <div className="table-scroll">
            <table>
              <thead>
                <tr>
                  <th>Дата и время</th>
                  {role !== "client" && <th>Клиент</th>}
                  <th>Описание обращения</th>
                  <th>Статус</th>
                  <th>
                    <span className="sr-only">Действия</span>
                  </th>
                </tr>
              </thead>
              <tbody>
                {visible.map((item) => (
                  <tr key={item.id}>
                    <td>
                      <strong>{formatDate(item.slot_start)}</strong>
                      <span className="cell-detail">
                        {formatTime(item.slot_start)} –{" "}
                        {formatTime(item.slot_end)}
                        <span className="subtle"> · 1 час</span>
                      </span>
                    </td>
                    {role !== "client" && (
                      <td>
                        {users.data?.find((user) => user.id === item.user_id)
                          ?.fio || `Клиент №${item.user_id}`}
                      </td>
                    )}
                    <td className="description-cell">
                      <span>{item.problem_description || "Без описания"}</span>
                      <small>Запись №{item.id}</small>
                    </td>
                    <td>
                      <span className={`status ${item.status}`}>
                        <i />
                        {statusLabels[item.status]}
                      </span>
                    </td>
                    <td className="row-action">
                      {item.status === "confirmed" &&
                        ((role === "client" &&
                          new Date(item.slot_start).getTime() > now) ||
                          (role === "staff" &&
                            new Date(item.slot_end).getTime() <= now)) && (
                          <button
                            className="text-button"
                            onClick={() => setConfirm(item)}
                          >
                            {role === "client" ? "Отменить" : "Завершить"}
                          </button>
                        )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
        <div className="panel-bottom">
          <TimezoneNote />
          <span className="muted">
            {visible.length} из {bookings.length}
          </span>
        </div>
      </section>
      <div className="info-banner">
        <span className="stat-icon blue">
          <CalendarDays size={20} />
        </span>
        <div>
          <strong>
            {role === "client" ? "Планы изменились?" : "Актуальное расписание"}
          </strong>
          <p>
            {role === "client"
              ? "Будущую запись можно отменить до начала визита."
              : "Статус записи обновляется после выполнения действия."}
          </p>
        </div>
      </div>
      {confirm && (
        <Confirm
          title={role === "client" ? "Отменить запись?" : "Завершить запись?"}
          label={role === "client" ? "Отменить запись" : "Завершить"}
          onClose={() => setConfirm(null)}
          onConfirm={async () => {
            if (role === "client") await api.cancelBooking(confirm.id);
            else await api.completeBooking(confirm.id);
            await cache.invalidateQueries({ queryKey: ["bookings"] });
          }}
        >
          Запись №{confirm.id} на {formatDateTime(confirm.slot_start)}.{" "}
          {role === "client"
            ? "Отменённая запись останется в истории."
            : "Запись будет отмечена как завершённая."}
        </Confirm>
      )}
    </>
  );
}

export function NewBookingPage() {
  const [date, setDate] = useState("");
  const [time, setTime] = useState("");
  const [description, setDescription] = useState("");
  const navigate = useNavigate();
  const cache = useQueryClient();
  const mutation = useMutation({
    mutationFn: async () => {
      const start = new Date(`${date}T${time}`);
      if (Number.isNaN(start.getTime()) || start.getTime() <= Date.now())
        throw new Error("Выберите будущие дату и время.");
      // Detect nonexistent local times during a daylight-saving transition.
      if (
        start.getHours() !== Number(time.split(":")[0]) ||
        start.getMinutes() !== Number(time.split(":")[1])
      )
        throw new Error(
          "Это местное время недоступно из-за перевода часов. Выберите другое время.",
        );
      return api.createBooking({
        slot_start: start.toISOString(),
        problem_description: description.trim(),
      });
    },
    onSuccess: async () => {
      await cache.invalidateQueries({ queryKey: ["bookings"] });
      navigate(roleHome("client"));
    },
  });
  const conflict =
    mutation.error instanceof ApiError &&
    (mutation.error.status === 409 ||
      (mutation.error.status === 400 &&
        mutation.error.message.includes("недоступны")));
  function submit(event: FormEvent) {
    event.preventDefault();
    if (!mutation.isPending) mutation.mutate();
  }
  return (
    <>
      <Link className="back-link" to="/client/bookings">
        <ArrowLeft size={16} />К моим записям
      </Link>
      <PageHeading
        eyebrow="ПЛАНИРУЕМ ВИЗИТ"
        title="Новая запись"
        description="Выберите дату, время и расскажите, с чем нужна помощь."
      />
      <div className="form-layout">
        <section className="panel form-panel">
          <h2>Детали записи</h2>
          <form className="stack" onSubmit={submit}>
            <fieldset disabled={mutation.isPending}>
              <div className="form-grid">
                <label>
                  Дата
                  <input
                    type="date"
                    required
                    value={date}
                    onChange={(event) => setDate(event.target.value)}
                  />
                </label>
                <label>
                  Время начала
                  <input
                    type="time"
                    required
                    step="60"
                    value={time}
                    onChange={(event) => setTime(event.target.value)}
                  />
                </label>
              </div>
              <TimezoneNote />
              <label>
                Описание обращения{" "}
                <span className="optional">необязательно</span>
                <textarea
                  value={description}
                  onChange={(event) => setDescription(event.target.value)}
                  rows={5}
                  placeholder="Кратко опишите ваш вопрос"
                />
              </label>
            </fieldset>
            <Feedback error={mutation.error} />
            {conflict && (
              <p className="alert warning" role="status">
                Это время уже занято. Выберите другое время — описание
                сохранено.
              </p>
            )}
            <div className="form-actions">
              <Link className="button secondary" to="/client/bookings">
                Назад
              </Link>
              <Submit pending={mutation.isPending}>
                Создать запись
                <ArrowRight size={16} />
              </Submit>
            </div>
          </form>
        </section>
        <aside className="booking-note">
          <span className="stat-icon blue">
            <Clock3 size={24} />
          </span>
          <h2>Один визит — один час</h2>
          <p>Доступность выбранного времени проверяется при создании записи.</p>
          <p>После подтверждения запись появится в вашем расписании.</p>
          {date && time && (
            <div className="booking-summary">
              <small>ВЫ ВЫБРАЛИ</small>
              <strong>
                {formatDate(new Date(`${date}T${time}`).toISOString())}
              </strong>
              <span>{time} · 1 час</span>
            </div>
          )}
        </aside>
      </div>
    </>
  );
}
