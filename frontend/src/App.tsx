import { Navigate, Route, Routes } from "react-router-dom";
import { ProtectedRoute, roleHome, useAuth } from "./auth";
import { Layout } from "./components/Layout";
import { Loading } from "./components/ui";
import { AuthPage } from "./pages/AuthPage";
import { BookingsPage, NewBookingPage } from "./pages/BookingsPage";
import { NotificationsPage } from "./pages/NotificationsPage";
import { ProfilePage } from "./pages/ProfilePage";
import { SecurityPage } from "./pages/SecurityPage";
import { UsersPage } from "./pages/UsersPage";

function Home() {
  const { user, loading } = useAuth();
  if (loading) return <Loading full />;
  return <Navigate to={user ? roleHome(user.role) : "/login"} replace />;
}
export function App() {
  return (
    <Routes>
      <Route path="/login" element={<AuthPage />} />
      <Route path="/register" element={<AuthPage register />} />
      <Route element={<ProtectedRoute />}>
        <Route element={<Layout />}>
          <Route path="/profile" element={<ProfilePage />} />
          <Route path="/security" element={<SecurityPage />} />
          <Route path="/notifications" element={<NotificationsPage />} />
          <Route element={<ProtectedRoute roles={["client"]} />}>
            <Route path="/client/bookings" element={<BookingsPage />} />
            <Route path="/client/bookings/new" element={<NewBookingPage />} />
          </Route>
          <Route element={<ProtectedRoute roles={["staff"]} />}>
            <Route path="/staff/bookings" element={<BookingsPage />} />
            <Route path="/staff/users" element={<UsersPage />} />
          </Route>
          <Route element={<ProtectedRoute roles={["admin"]} />}>
            <Route path="/admin/bookings" element={<BookingsPage />} />
            <Route path="/admin/users" element={<UsersPage />} />
          </Route>
        </Route>
      </Route>
      <Route path="/" element={<Home />} />
      <Route
        path="*"
        element={
          <div className="standalone">
            <h1>Страница не найдена</h1>
            <p className="muted">
              Проверьте адрес или вернитесь в личный кабинет.
            </p>
            <a className="button" href="/">
              В личный кабинет
            </a>
          </div>
        }
      />
    </Routes>
  );
}
