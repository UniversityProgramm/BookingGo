import { useQuery } from "@tanstack/react-query";
import { Bell, RefreshCw } from "lucide-react";
import { api } from "../api";
import { useAuth } from "../auth";
import {
  Empty,
  ErrorState,
  formatDateTime,
  Loading,
  PageHeading,
  TimezoneNote,
} from "../components/ui";

export function NotificationsPage() {
  const { user } = useAuth();
  // This GET marks messages as read: never prefetch it or refetch on focus/reconnect.
  const query = useQuery({
    queryKey: ["notifications", user!.id],
    queryFn: ({ signal }) => api.notifications(signal),
    retry: false,
    refetchOnWindowFocus: false,
    refetchOnReconnect: false,
    refetchOnMount: "always",
    gcTime: 0,
  });
  const notifications = [...(query.data?.notifications ?? [])].sort((a, b) =>
    b.created_at.localeCompare(a.created_at),
  );
  return (
    <>
      <PageHeading
        eyebrow="ОСТАВАЙТЕСЬ В КУРСЕ"
        title="Уведомления"
        description="Изменения в записях и важные события вашего аккаунта."
        action={
          <button
            className="button secondary"
            disabled={query.isFetching}
            onClick={() => {
              void query.refetch();
            }}
          >
            <RefreshCw size={16} className={query.isFetching ? "spin" : ""} />
            Обновить
          </button>
        }
      />
      <p className="small muted">
        При открытии этого раздела уведомления помечаются прочитанными.
      </p>
      {query.data?.warning && (
        <p className="alert warning" role="status">
          {query.data.warning}
        </p>
      )}
      <section className="panel">
        {query.isPending ? (
          <Loading />
        ) : query.error ? (
          <ErrorState
            error={query.error}
            retry={() => {
              void query.refetch();
            }}
          />
        ) : !notifications.length ? (
          <Empty title="Вы всё прочитали">
            Новые уведомления о записях появятся здесь.
          </Empty>
        ) : (
          <div className="notification-list">
            {notifications.map((item) => (
              <article key={item.id} className="notification">
                <span className="stat-icon blue">
                  <Bell size={20} />
                </span>
                <div>
                  <h2>{item.title}</h2>
                  <p>{item.body}</p>
                  <time dateTime={item.created_at}>
                    {formatDateTime(item.created_at)}
                  </time>
                </div>
              </article>
            ))}
          </div>
        )}
        <div className="panel-bottom">
          <TimezoneNote />
        </div>
      </section>
    </>
  );
}
