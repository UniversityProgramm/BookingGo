import { useState, type FormEvent } from "react";
import { Link, Navigate } from "react-router-dom";
import { ArrowRight, CalendarCheck2, Check, ShieldCheck } from "lucide-react";
import { useMutation } from "@tanstack/react-query";
import { api } from "../api";
import { roleHome, useAuth } from "../auth";
import { ErrorState, Feedback, Loading, Submit } from "../components/ui";

export function AuthPage({ register = false }: { register?: boolean }) {
  const auth = useAuth();
  const [form, setForm] = useState({
    email: "",
    password: "",
    fio: "",
    phone: "",
  });
  const mutation = useMutation({
    mutationFn: () =>
      register
        ? api.register(form)
        : api.login({ email: form.email, password: form.password }),
    onSuccess: (result) => auth.establish(result.token),
  });
  function submit(event: FormEvent) {
    event.preventDefault();
    if (!mutation.isPending) mutation.mutate();
  }
  if (auth.loading) return <Loading full />;
  if (auth.error)
    return (
      <div className="standalone">
        <ErrorState error={auth.error} retry={auth.retry} />
        <button className="button secondary" onClick={() => auth.end()}>
          Перейти ко входу
        </button>
      </div>
    );
  if (auth.user) return <Navigate to={roleHome(auth.user.role)} replace />;
  const update = (key: keyof typeof form, value: string) =>
    setForm((previous) => ({ ...previous, [key]: value }));
  return (
    <main className="auth-page">
      <section className="auth-story">
        <Link className="brand" to="/login">
          <span className="brand-mark">
            <CalendarCheck2 size={25} />
          </span>
          Booking<span>Go</span>
        </Link>
        <div className="auth-pitch">
          <p className="eyebrow">ВАШЕ ВРЕМЯ ПОД КОНТРОЛЕМ</p>
          <h1>
            Меньше хлопот.
            <br />
            Больше планов.
          </h1>
          <p>
            Выбирайте удобное время, управляйте записями и будьте в курсе
            изменений.
          </p>
          <div className="story-card">
            <span className="story-check">
              <Check size={24} />
            </span>
            <div>
              <strong>Всё в одном месте</strong>
              <p>Записи, уведомления и ваш профиль</p>
            </div>
          </div>
        </div>
        <p className="auth-foot">
          <ShieldCheck size={16} />
          Простой и безопасный личный кабинет
        </p>
      </section>
      <section className="auth-form-panel">
        <div className="auth-form">
          <p className="eyebrow">ЛИЧНЫЙ КАБИНЕТ</p>
          <h2>{register ? "Давайте знакомиться" : "С возвращением"}</h2>
          <p className="muted">
            {register
              ? "Создайте аккаунт, чтобы записаться на удобное время."
              : "Войдите, чтобы управлять своими записями."}
          </p>
          <Feedback success={auth.notice} />
          <form onSubmit={submit} className="stack">
            <fieldset disabled={mutation.isPending}>
              {register && (
                <>
                  <label>
                    ФИО
                    <input
                      autoComplete="name"
                      value={form.fio}
                      onChange={(event) => update("fio", event.target.value)}
                      required
                      placeholder="Иванов Иван Иванович"
                    />
                  </label>
                  <label>
                    Телефон
                    <input
                      type="tel"
                      autoComplete="tel"
                      minLength={10}
                      maxLength={15}
                      value={form.phone}
                      onChange={(event) => update("phone", event.target.value)}
                      required
                      placeholder="+7 999 123-45-67"
                    />
                  </label>
                </>
              )}
              <label>
                Электронная почта
                <input
                  type="email"
                  autoComplete="email"
                  value={form.email}
                  onChange={(event) => update("email", event.target.value)}
                  required
                  placeholder="you@example.com"
                />
              </label>
              <label>
                Пароль
                <input
                  type="password"
                  autoComplete={register ? "new-password" : "current-password"}
                  minLength={8}
                  maxLength={20}
                  value={form.password}
                  onChange={(event) => update("password", event.target.value)}
                  required
                  placeholder="От 8 до 20 символов"
                />
              </label>
            </fieldset>
            <Feedback error={mutation.error} />
            <Submit pending={mutation.isPending}>
              {register ? "Создать аккаунт" : "Войти"}
              <ArrowRight size={17} />
            </Submit>
          </form>
          <p className="auth-switch">
            {register ? "Уже есть аккаунт?" : "Ещё нет аккаунта?"}{" "}
            <Link to={register ? "/login" : "/register"}>
              {register ? "Войти" : "Зарегистрироваться"}
            </Link>
          </p>
        </div>
      </section>
    </main>
  );
}
