import { useState, type FormEvent } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Plus, Search } from "lucide-react";
import { api } from "../api";
import { roleLabels, useAuth } from "../auth";
import type { User } from "../types";
import {
  Confirm,
  Empty,
  ErrorState,
  Feedback,
  formatDate,
  Loading,
  Modal,
  PageHeading,
  Submit,
} from "../components/ui";

export function UsersPage() {
  const { user } = useAuth();
  const role = user!.role;
  const cache = useQueryClient();
  const query = useQuery({
    queryKey: ["users", role],
    queryFn: ({ signal }) => api.users(role, signal),
  });
  const [search, setSearch] = useState("");
  const [selected, setSelected] = useState<User | "new" | null>(null);
  const [deleting, setDeleting] = useState<User | null>(null);
  const visible = (query.data ?? []).filter((item) =>
    `${item.id} ${item.fio} ${item.email} ${item.phone}`
      .toLocaleLowerCase("ru")
      .includes(search.toLocaleLowerCase("ru")),
  );
  return (
    <>
      <PageHeading
        eyebrow="КЛИЕНТЫ И КОМАНДА"
        title="Пользователи"
        description={
          role === "admin"
            ? "Контактные данные и управление аккаунтами."
            : "Контактные данные клиентов и команды."
        }
        action={
          role === "admin" && (
            <button className="button" onClick={() => setSelected("new")}>
              <Plus size={18} />
              Создать клиента
            </button>
          )
        }
      />
      <section className="panel">
        <div className="panel-title">
          <h2>Все пользователи</h2>
          <span className="count">{query.data?.length ?? 0}</span>
        </div>
        <div className="toolbar">
          <label className="search wide">
            <Search size={18} />
            <input
              aria-label="Поиск пользователей"
              value={search}
              onChange={(event) => setSearch(event.target.value)}
              placeholder="Имя, email, телефон или номер"
            />
          </label>
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
        ) : !visible.length ? (
          <Empty title="Пользователи не найдены">
            Попробуйте изменить поисковый запрос.
          </Empty>
        ) : (
          <div className="table-scroll">
            <table>
              <thead>
                <tr>
                  <th>Пользователь</th>
                  <th>Контакты</th>
                  <th>Роль</th>
                  <th>Действия</th>
                </tr>
              </thead>
              <tbody>
                {visible.map((item) => (
                  <tr key={item.id}>
                    <td>
                      <strong>{item.fio || "Без имени"}</strong>
                      <small className="cell-detail">
                        №{item.id} · с {formatDate(item.created_at)}
                      </small>
                    </td>
                    <td>
                      <span>{item.email}</span>
                      <small className="cell-detail">{item.phone}</small>
                    </td>
                    <td>
                      <span className="role-pill">{roleLabels[item.role]}</span>
                    </td>
                    <td>
                      <div className="inline-actions">
                        <button
                          className="text-button"
                          onClick={() => setSelected(item)}
                        >
                          {role === "admin" ? "Изменить" : "Подробнее"}
                        </button>
                        {role === "admin" && (
                          <button
                            className="text-button danger-text"
                            onClick={() => setDeleting(item)}
                          >
                            Удалить
                          </button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>
      {selected && (
        <UserEditor
          key={selected === "new" ? "new" : selected.id}
          selected={selected}
          readOnly={role !== "admin"}
          onClose={() => setSelected(null)}
        />
      )}
      {deleting && (
        <Confirm
          title="Удалить пользователя?"
          label="Удалить"
          onClose={() => setDeleting(null)}
          onConfirm={async () => {
            await api.deleteUser(deleting.id);
            await cache.invalidateQueries({ queryKey: ["users"] });
          }}
        >
          Аккаунт {deleting.fio || deleting.email} будет удалён. Это действие
          нельзя отменить.
        </Confirm>
      )}
    </>
  );
}
function UserEditor({
  selected,
  readOnly,
  onClose,
}: {
  selected: User | "new";
  readOnly: boolean;
  onClose: () => void;
}) {
  const creating = selected === "new";
  const cache = useQueryClient();
  const [form, setForm] = useState({
    fio: creating ? "" : selected.fio,
    phone: creating ? "" : selected.phone,
    email: creating ? "" : selected.email,
    password: "",
  });
  const mutation = useMutation({
    mutationFn: () =>
      creating
        ? api.createUser(form)
        : api.updateUser(selected.id, { fio: form.fio, phone: form.phone }),
    onSuccess: async () => {
      await cache.invalidateQueries({ queryKey: ["users"] });
      onClose();
    },
  });
  function submit(event: FormEvent) {
    event.preventDefault();
    if (!mutation.isPending && !readOnly) mutation.mutate();
  }
  const field = (key: keyof typeof form, value: string) =>
    setForm((previous) => ({ ...previous, [key]: value }));
  return (
    <Modal
      title={
        creating
          ? "Новый клиент"
          : readOnly
            ? "Данные пользователя"
            : "Изменить пользователя"
      }
      onClose={onClose}
      busy={mutation.isPending}
    >
      <form onSubmit={submit} className="stack">
        <fieldset disabled={mutation.isPending}>
          <label>
            ФИО
            <input
              value={form.fio}
              readOnly={readOnly}
              onChange={(event) => field("fio", event.target.value)}
              required
              autoComplete="name"
            />
          </label>
          <label>
            Телефон
            <input
              type="tel"
              value={form.phone}
              readOnly={readOnly}
              onChange={(event) => field("phone", event.target.value)}
              required
              minLength={10}
              maxLength={15}
              autoComplete="tel"
            />
          </label>
          <label>
            Электронная почта
            <input
              type="email"
              value={form.email}
              readOnly={!creating}
              onChange={(event) => field("email", event.target.value)}
              required
              autoComplete="email"
            />
          </label>
          {creating && (
            <label>
              Пароль
              <input
                type="password"
                required
                minLength={8}
                maxLength={20}
                value={form.password}
                onChange={(event) => field("password", event.target.value)}
                autoComplete="new-password"
              />
            </label>
          )}
        </fieldset>
        <Feedback error={mutation.error} />
        <div className="form-actions">
          <button
            type="button"
            className="button secondary"
            disabled={mutation.isPending}
            onClick={onClose}
          >
            Закрыть
          </button>
          {!readOnly && (
            <Submit pending={mutation.isPending}>
              {creating ? "Создать клиента" : "Сохранить изменения"}
            </Submit>
          )}
        </div>
      </form>
    </Modal>
  );
}
