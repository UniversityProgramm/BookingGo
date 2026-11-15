import { useEffect, useState, type FormEvent } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { KeyRound, LockKeyhole, Mail, ShieldCheck } from "lucide-react";
import QRCode from "qrcode";
import { api } from "../api";
import { useAuth } from "../auth";
import { Feedback, PageHeading, Submit } from "../components/ui";

export function SecurityPage() {
  const auth = useAuth();
  const enabled = auth.user!.is_totp_enabled;
  const [email, setEmail] = useState({
    new_email: "",
    confirm_password: "",
    otp_code: "",
  });
  const [password, setPassword] = useState({
    cur_password: "",
    new_password: "",
    otp_code: "",
  });
  const emailMutation = useMutation({
    mutationFn: () =>
      api.changeEmail({
        ...email,
        otp_code: enabled ? email.otp_code : undefined,
      }),
    onSuccess: () => auth.end("Почта изменена. Войдите с новым адресом."),
  });
  const passwordMutation = useMutation({
    mutationFn: () =>
      api.changePassword({
        ...password,
        otp_code: enabled ? password.otp_code : undefined,
      }),
    onSuccess: () => auth.end("Пароль изменён. Войдите с новым паролем."),
  });
  const saveEmail = (event: FormEvent) => {
    event.preventDefault();
    if (!emailMutation.isPending) emailMutation.mutate();
  };
  const savePassword = (event: FormEvent) => {
    event.preventDefault();
    if (!passwordMutation.isPending) passwordMutation.mutate();
  };
  return (
    <>
      <PageHeading
        eyebrow="ДОСТУП К АККАУНТУ"
        title="Безопасность"
        description="Управляйте данными для входа и защитой важных действий."
      />
      <div className="security-grid">
        <section className="panel form-panel">
          <div className="section-heading">
            <Mail size={21} />
            <h2>Изменить почту</h2>
          </div>
          <p className="muted">Сейчас: {auth.user!.email}</p>
          <form className="stack" onSubmit={saveEmail}>
            <fieldset disabled={emailMutation.isPending}>
              <label>
                Новая электронная почта
                <input
                  type="email"
                  autoComplete="email"
                  value={email.new_email}
                  onChange={(event) =>
                    setEmail({ ...email, new_email: event.target.value })
                  }
                  required
                />
              </label>
              <label>
                Текущий пароль
                <input
                  type="password"
                  autoComplete="current-password"
                  value={email.confirm_password}
                  onChange={(event) =>
                    setEmail({ ...email, confirm_password: event.target.value })
                  }
                  required
                />
              </label>
              {enabled && (
                <OtpInput
                  value={email.otp_code}
                  onChange={(value) => setEmail({ ...email, otp_code: value })}
                />
              )}
            </fieldset>
            <Feedback error={emailMutation.error} />
            <div>
              <Submit pending={emailMutation.isPending}>Изменить почту</Submit>
            </div>
          </form>
        </section>
        <section className="panel form-panel">
          <div className="section-heading">
            <LockKeyhole size={21} />
            <h2>Изменить пароль</h2>
          </div>
          <p className="muted">После изменения потребуется снова войти.</p>
          <form className="stack" onSubmit={savePassword}>
            <fieldset disabled={passwordMutation.isPending}>
              <label>
                Текущий пароль
                <input
                  type="password"
                  autoComplete="current-password"
                  value={password.cur_password}
                  onChange={(event) =>
                    setPassword({
                      ...password,
                      cur_password: event.target.value,
                    })
                  }
                  required
                />
              </label>
              <label>
                Новый пароль
                <input
                  type="password"
                  autoComplete="new-password"
                  minLength={8}
                  maxLength={20}
                  value={password.new_password}
                  onChange={(event) =>
                    setPassword({
                      ...password,
                      new_password: event.target.value,
                    })
                  }
                  required
                />
                <small className="muted">От 8 до 20 символов</small>
              </label>
              {enabled && (
                <OtpInput
                  value={password.otp_code}
                  onChange={(value) =>
                    setPassword({ ...password, otp_code: value })
                  }
                />
              )}
            </fieldset>
            <Feedback error={passwordMutation.error} />
            <div>
              <Submit pending={passwordMutation.isPending}>
                Изменить пароль
              </Submit>
            </div>
          </form>
        </section>
      </div>
      <TotpPanel />
    </>
  );
}
function OtpInput({
  value,
  onChange,
}: {
  value: string;
  onChange: (value: string) => void;
}) {
  return (
    <label>
      Код из приложения-аутентификатора
      <input
        type="text"
        inputMode="numeric"
        pattern="[0-9]{6}"
        maxLength={6}
        autoComplete="one-time-code"
        value={value}
        onChange={(event) => onChange(event.target.value.replace(/\D/g, ""))}
        required
        placeholder="000000"
      />
    </label>
  );
}
function TotpPanel() {
  const { user } = useAuth();
  const enabled = user!.is_totp_enabled;
  const cache = useQueryClient();
  const [url, setUrl] = useState("");
  const [qr, setQr] = useState("");
  const [qrError, setQrError] = useState<Error | null>(null);
  const [code, setCode] = useState("");
  const [password, setPassword] = useState("");
  const [disable, setDisable] = useState(false);
  const setup = useMutation({
    mutationFn: api.setupTotp,
    onSuccess: (data) => {
      setUrl(data.manual_entry_key);
      setCode("");
    },
  });
  const verify = useMutation({
    mutationFn: () => api.verifyTotp(code),
    onSuccess: async () => {
      setUrl("");
      setQr("");
      setCode("");
      await cache.invalidateQueries({ queryKey: ["profile"] });
    },
  });
  const disableMutation = useMutation({
    mutationFn: () => api.disableTotp(code, password),
    onSuccess: async () => {
      setPassword("");
      setCode("");
      setDisable(false);
      await cache.invalidateQueries({ queryKey: ["profile"] });
    },
  });
  useEffect(() => {
    let active = true;
    setQr("");
    setQrError(null);
    if (url)
      QRCode.toDataURL(url, {
        width: 220,
        margin: 2,
        errorCorrectionLevel: "M",
      })
        .then((value) => {
          if (active) setQr(value);
        })
        .catch(() => {
          if (active)
            setQrError(
              new Error(
                "Не удалось создать QR-код. Используйте ключ для ручного ввода.",
              ),
            );
        });
    return () => {
      active = false;
    };
  }, [url]);
  const secret = url ? new URL(url).searchParams.get("secret") : "";
  const submit = (event: FormEvent) => {
    event.preventDefault();
    if (disable ? !disableMutation.isPending : !verify.isPending) {
      if (disable) disableMutation.mutate();
      else verify.mutate();
    }
  };
  return (
    <section className="panel form-panel totp-panel">
      <div className="panel-title flush">
        <div className="section-heading">
          <ShieldCheck size={22} />
          <h2>Двухфакторная защита</h2>
        </div>
        <span className={`status ${enabled ? "confirmed" : "cancelled"}`}>
          <i />
          {enabled ? "Включена" : "Не включена"}
        </span>
      </div>
      <p className="muted">
        Подтверждайте смену почты и пароля одноразовым кодом из
        приложения-аутентификатора.
      </p>
      <Feedback
        error={setup.error}
        success={
          verify.isSuccess && enabled
            ? "Двухфакторная защита включена."
            : disableMutation.isSuccess && !enabled
              ? "Двухфакторная защита отключена."
              : ""
        }
      />
      {!enabled && !url && (
        <button
          className="button secondary"
          onClick={() => {
            if (!setup.isPending) setup.mutate();
          }}
          disabled={setup.isPending}
        >
          <KeyRound size={17} />
          {setup.isPending ? "Подготовка…" : "Настроить защиту"}
        </button>
      )}
      {!enabled && url && (
        <div className="totp-setup">
          <div className="qr-block">
            {qr ? (
              <img
                src={qr}
                width="220"
                height="220"
                alt="QR-код для настройки TOTP в приложении-аутентификаторе"
              />
            ) : (
              <p role="status">Создаём QR-код…</p>
            )}
          </div>
          <div>
            <h3>1. Добавьте аккаунт в приложение</h3>
            <p className="muted">
              Отсканируйте QR-код или введите ключ вручную:
            </p>
            <code className="totp-secret">{secret}</code>
            <Feedback error={qrError} />
            <h3>2. Подтвердите настройку</h3>
            <form className="stack" onSubmit={submit}>
              <fieldset disabled={verify.isPending}>
                <OtpInput value={code} onChange={setCode} />
              </fieldset>
              <Feedback error={verify.error} />
              <div className="form-actions">
                <button
                  type="button"
                  className="button secondary"
                  disabled={verify.isPending}
                  onClick={() => {
                    setUrl("");
                    setCode("");
                    verify.reset();
                  }}
                >
                  Отмена
                </button>
                <Submit pending={verify.isPending}>Включить защиту</Submit>
              </div>
            </form>
          </div>
        </div>
      )}
      {enabled && !disable && (
        <button
          className="button secondary"
          onClick={() => {
            setDisable(true);
            setCode("");
            disableMutation.reset();
          }}
        >
          Отключить защиту
        </button>
      )}
      {enabled && disable && (
        <form className="stack narrow" onSubmit={submit}>
          <fieldset disabled={disableMutation.isPending}>
            <label>
              Текущий пароль
              <input
                type="password"
                autoComplete="current-password"
                required
                value={password}
                onChange={(event) => setPassword(event.target.value)}
              />
            </label>
            <OtpInput value={code} onChange={setCode} />
          </fieldset>
          <Feedback error={disableMutation.error} />
          <div className="form-actions">
            <button
              type="button"
              className="button secondary"
              disabled={disableMutation.isPending}
              onClick={() => {
                setDisable(false);
                setCode("");
                setPassword("");
              }}
            >
              Отмена
            </button>
            <Submit pending={disableMutation.isPending}>
              Отключить защиту
            </Submit>
          </div>
        </form>
      )}
    </section>
  );
}
