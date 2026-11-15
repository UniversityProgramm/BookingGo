import { useEffect, useState } from "react";
import { Link, NavLink, Outlet, useLocation } from "react-router-dom";
import {
  Bell,
  CalendarCheck2,
  CalendarDays,
  ChevronRight,
  LogOut,
  Menu,
  ShieldCheck,
  UserRound,
  Users,
  X,
} from "lucide-react";
import { api } from "../api";
import { roleHome, roleLabels, useAuth } from "../auth";
import { Confirm } from "./ui";

export function Layout() {
  const { user, end } = useAuth();
  const location = useLocation();
  const [open, setOpen] = useState(false);
  const [logout, setLogout] = useState(false);
  useEffect(() => {
    setOpen(false);
    document.getElementById("main-content")?.focus();
  }, [location.pathname]);
  if (!user) return null;
  const nav = [
    {
      to: roleHome(user.role),
      label: user.role === "client" ? "Мои записи" : "Все записи",
      icon: CalendarDays,
    },
    ...(user.role === "client"
      ? []
      : [{ to: `/${user.role}/users`, label: "Пользователи", icon: Users }]),
    { to: "/notifications", label: "Уведомления", icon: Bell },
    { to: "/profile", label: "Мой профиль", icon: UserRound },
    { to: "/security", label: "Безопасность", icon: ShieldCheck },
  ];
  const current =
    nav.find((item) => location.pathname.startsWith(item.to))?.label ??
    "Новая запись";
  return (
    <div className="app-shell">
      <a className="skip-link" href="#main-content">
        К содержимому
      </a>
      <header className="mobile-header">
        <Link className="brand" to={roleHome(user.role)}>
          <span className="brand-mark">
            <CalendarCheck2 size={23} />
          </span>
          Booking<span>Go</span>
        </Link>
        <button
          className="icon-button"
          aria-label={open ? "Закрыть меню" : "Открыть меню"}
          aria-expanded={open}
          aria-controls="sidebar"
          onClick={() => setOpen(!open)}
        >
          {open ? <X /> : <Menu />}
        </button>
      </header>
      {open && (
        <button
          className="nav-backdrop"
          aria-label="Закрыть меню"
          onClick={() => setOpen(false)}
        />
      )}
      <aside className={`sidebar ${open ? "is-open" : ""}`} id="sidebar">
        <Link className="brand" to={roleHome(user.role)}>
          <span className="brand-mark">
            <CalendarCheck2 size={25} />
          </span>
          Booking<span>Go</span>
        </Link>
        <p className="nav-caption">ЛИЧНЫЙ КАБИНЕТ</p>
        <nav aria-label="Основная навигация">
          {nav.map(({ to, label, icon: Icon }) => (
            <NavLink to={to} key={to}>
              <Icon size={20} />
              <span>{label}</span>
            </NavLink>
          ))}
        </nav>
        <div className="sidebar-bottom">
          <div className="help-card">
            <span className="help-dot" />
            <div>
              <strong>Всё по плану</strong>
              <p>Ваши записи всегда под рукой</p>
            </div>
          </div>
          <Link to="/profile" className="account">
            <span className="avatar">
              {user.fio.trim().charAt(0).toUpperCase() || "Б"}
            </span>
            <span>
              <strong>{user.fio || user.email}</strong>
              <small>{roleLabels[user.role]}</small>
            </span>
          </Link>
          <button className="logout" onClick={() => setLogout(true)}>
            <LogOut size={17} />
            Выйти из аккаунта
          </button>
        </div>
      </aside>
      <div className="workspace">
        <div className="topbar">
          <span>
            Личный кабинет
            <ChevronRight size={14} />
            <strong>{current}</strong>
          </span>
          <span className="role-pill">{roleLabels[user.role]}</span>
        </div>
        <main id="main-content" tabIndex={-1}>
          <Outlet />
        </main>
        <footer className="workspace-footer">
          <span>BookingGo</span>
          <span>Время для того, что важно</span>
        </footer>
      </div>
      {logout && (
        <Confirm
          title="Выйти из аккаунта?"
          label="Выйти"
          onClose={() => setLogout(false)}
          onConfirm={async () => {
            await api.logout();
            end("Вы вышли из аккаунта.");
          }}
        >
          Для доступа к записям потребуется снова войти.
        </Confirm>
      )}
    </div>
  );
}
