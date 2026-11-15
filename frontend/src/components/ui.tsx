import {
  useEffect,
  useRef,
  useState,
  type FormEvent,
  type ReactNode,
} from "react";
import {
  AlertCircle,
  CalendarDays,
  LoaderCircle,
  RefreshCw,
  X,
} from "lucide-react";

export function Loading({ full = false }: { full?: boolean }) {
  return (
    <div className={`loading ${full ? "full" : ""}`} role="status">
      <LoaderCircle className="spin" size={22} /> Загрузка…
    </div>
  );
}
export function ErrorState({
  error,
  retry,
}: {
  error: Error;
  retry: () => void;
}) {
  return (
    <div className="state error-state" role="alert">
      <AlertCircle size={28} />
      <h3>Не удалось загрузить данные</h3>
      <p>{error.message}</p>
      <button className="button secondary" onClick={retry}>
        <RefreshCw size={16} />
        Повторить попытку
      </button>
    </div>
  );
}
export function Empty({
  title = "Пока ничего нет",
  children,
}: {
  title?: string;
  children?: ReactNode;
}) {
  return (
    <div className="state">
      <span className="empty-icon">
        <CalendarDays size={28} />
      </span>
      <h3>{title}</h3>
      <p>{children}</p>
    </div>
  );
}
export function Feedback({
  error,
  success,
}: {
  error?: Error | null;
  success?: string;
}) {
  return (
    <>
      {error && (
        <p className="alert danger" role="alert">
          {error.message}
        </p>
      )}
      {success && (
        <p className="alert success" role="status">
          {success}
        </p>
      )}
    </>
  );
}
export function PageHeading({
  eyebrow,
  title,
  description,
  action,
}: {
  eyebrow: string;
  title: string;
  description: string;
  action?: ReactNode;
}) {
  return (
    <header className="page-heading">
      <div>
        <p className="eyebrow">{eyebrow}</p>
        <h1>{title}</h1>
        <p className="muted">{description}</p>
      </div>
      {action}
    </header>
  );
}
export function Submit({
  pending,
  children,
}: {
  pending: boolean;
  children: ReactNode;
}) {
  return (
    <button className="button" type="submit" disabled={pending}>
      {pending && <LoaderCircle className="spin" size={16} />}
      {pending ? "Подождите…" : children}
    </button>
  );
}
export function Modal({
  title,
  children,
  onClose,
  busy = false,
}: {
  title: string;
  children: ReactNode;
  onClose: () => void;
  busy?: boolean;
}) {
  const dialog = useRef<HTMLDialogElement>(null);
  useEffect(() => {
    const previous = document.activeElement as HTMLElement | null;
    dialog.current?.showModal();
    return () => {
      previous?.focus();
    };
  }, []);
  return (
    <dialog
      ref={dialog}
      className="modal"
      aria-label={title}
      onCancel={(event) => {
        event.preventDefault();
        if (!busy) onClose();
      }}
    >
      <div className="modal-heading">
        <h2>{title}</h2>
        <button
          className="icon-button"
          aria-label="Закрыть"
          disabled={busy}
          onClick={onClose}
        >
          <X size={20} />
        </button>
      </div>
      {children}
    </dialog>
  );
}
export function Confirm({
  title,
  children,
  label,
  onClose,
  onConfirm,
}: {
  title: string;
  children: ReactNode;
  label: string;
  onClose: () => void;
  onConfirm: () => Promise<unknown>;
}) {
  const [pending, setPending] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const lock = useRef(false);
  async function submit(event: FormEvent) {
    event.preventDefault();
    if (lock.current) return;
    lock.current = true;
    setPending(true);
    setError(null);
    try {
      await onConfirm();
      onClose();
    } catch (error) {
      setError(error as Error);
    } finally {
      lock.current = false;
      setPending(false);
    }
  }
  return (
    <Modal title={title} onClose={onClose} busy={pending}>
      <form onSubmit={submit}>
        <p className="muted">{children}</p>
        <Feedback error={error} />
        <div className="form-actions">
          <button
            type="button"
            className="button secondary"
            onClick={onClose}
            disabled={pending}
          >
            Назад
          </button>
          <Submit pending={pending}>{label}</Submit>
        </div>
      </form>
    </Modal>
  );
}
export const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone;
export const formatDate = (value: string) =>
  new Intl.DateTimeFormat("ru-RU", {
    day: "numeric",
    month: "long",
    year: "numeric",
  }).format(new Date(value));
export const formatTime = (value: string) =>
  new Intl.DateTimeFormat("ru-RU", {
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(value));
export const formatDateTime = (value: string) =>
  `${formatDate(value)}, ${formatTime(value)}`;
export function TimezoneNote() {
  return (
    <p className="timezone">Время указано в вашем часовом поясе · {timezone}</p>
  );
}
