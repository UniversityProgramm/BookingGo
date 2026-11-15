import {
  createContext,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { Navigate, Outlet, useLocation } from "react-router-dom";
import { api, SESSION_EXPIRED, tokenStore } from "./api";
import type { Role, User } from "./types";
import { ErrorState, Loading } from "./components/ui";

export const roleLabels: Record<Role, string> = {
  client: "Клиент",
  staff: "Сотрудник",
  admin: "Администратор",
};
export const roleHome = (role: Role) => `/${role}/bookings`;
export const canAccess = (role: Role, roles?: readonly Role[]) =>
  !roles || roles.includes(role);
interface AuthContextValue {
  user: User | undefined;
  loading: boolean;
  error: Error | null;
  notice: string;
  establish: (token: string) => void;
  end: (notice?: string) => void;
  retry: () => void;
}
const AuthContext = createContext<AuthContextValue | null>(null);
export function AuthProvider({ children }: { children: ReactNode }) {
  const cache = useQueryClient();
  const [token, setToken] = useState(tokenStore.get);
  const [notice, setNotice] = useState("");
  const profile = useQuery({
    queryKey: ["profile", token],
    queryFn: ({ signal }) => api.me(signal),
    enabled: !!token,
    retry: false,
    staleTime: 60_000,
  });
  function end(message = "") {
    tokenStore.clear();
    setToken(null);
    cache.clear();
    setNotice(message);
  }
  useEffect(() => {
    const expired = () => end("Сессия завершена. Войдите заново.");
    window.addEventListener(SESSION_EXPIRED, expired);
    return () => window.removeEventListener(SESSION_EXPIRED, expired);
  }, [cache]);
  function establish(value: string) {
    cache.clear();
    tokenStore.set(value);
    setToken(value);
    setNotice("");
  }
  return (
    <AuthContext.Provider
      value={{
        user: token ? profile.data : undefined,
        loading: !!token && profile.isPending,
        error: token ? profile.error : null,
        notice,
        establish,
        end,
        retry: () => {
          void profile.refetch();
        },
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}
export function useAuth() {
  const value = useContext(AuthContext);
  if (!value) throw new Error("AuthProvider is required");
  return value;
}
export function ProtectedRoute({ roles }: { roles?: readonly Role[] }) {
  const auth = useAuth();
  const location = useLocation();
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
  if (!auth.user)
    return <Navigate to="/login" replace state={{ from: location.pathname }} />;
  if (!canAccess(auth.user.role, roles))
    return (
      <div className="standalone">
        <h1>Нет доступа</h1>
        <p>Этот раздел недоступен для вашей роли.</p>
        <a className="button" href={roleHome(auth.user.role)}>
          Мои записи
        </a>
      </div>
    );
  return <Outlet />;
}
