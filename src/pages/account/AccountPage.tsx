import { Link } from "react-router-dom";
import { resolveProfileAvatarUrl } from "../../entities/profile-avatar";
import { getActiveProfile } from "../../shared/lib/active-profile";
import { AppLogo } from "../../shared/ui/AppLogo";
import "./AccountPage.css";

const menuItems = [
  { icon: "hgi-user", label: "Профиль", active: true },
  { icon: "hgi-message-01", label: "Сообщения", badge: "3" },
  { icon: "hgi-favourite", label: "Избранное" },
  { icon: "hgi-wallet-02", label: "Кошелёк" },
  { icon: "hgi-settings-02", label: "Настройки" },
];

export function AccountPage() {
  const activeProfile = getActiveProfile();
  const activeAvatarUrl = resolveProfileAvatarUrl(activeProfile.profileCode, activeProfile.avatarUrl);

  return (
    <div className="account-page">
      <header className="account-header">
        <div className="account-header__inner">
          <Link to="/" className="account-header__brand" aria-label="На Авито">
            <AppLogo compact />
          </Link>

          <div className="account-header__actions">
            <Link to="/" className="account-header__market-link">Вернуться на Авито</Link>
            <button type="button" className="account-header__icon" aria-label="Сообщения">
              <i className="hgi hgi-stroke hgi-message-01" aria-hidden="true" />
              <span className="account-header__dot" aria-hidden="true" />
            </button>
            <button type="button" className="account-header__icon" aria-label="Уведомления">
              <i className="hgi hgi-stroke hgi-notification-02" aria-hidden="true" />
            </button>
            <Link
              to="/profiles?return=/account"
              className="account-header__avatar-link"
              aria-label={`Сменить аккаунт. Сейчас ${activeProfile.name}`}
              title="Сменить аккаунт"
            >
              <img className="account-header__avatar" src={activeAvatarUrl} alt="" />
              <span className="account-header__switch-dot" aria-hidden="true">
                <i className="hgi hgi-stroke hgi-arrow-right-01" />
              </span>
            </Link>
          </div>
        </div>
      </header>

      <main className="account-main">
        <aside className="account-sidebar">
          <Link to="/profiles?return=/account" className="account-profile-mini" aria-label="Сменить аккаунт">
            <img className="account-profile-mini__avatar" src={activeAvatarUrl} alt="" />
            <div>
              <strong>{activeProfile.name}</strong>
              <span>Сменить аккаунт</span>
            </div>
            <i className="hgi hgi-stroke hgi-arrow-right-01 account-profile-mini__arrow" aria-hidden="true" />
          </Link>

          <nav className="account-menu" aria-label="Разделы личного кабинета">
            {menuItems.map((item) => (
              <button key={item.label} type="button" className={`account-menu__item${item.active ? " is-active" : ""}`}>
                <i className={`hgi hgi-stroke ${item.icon}`} aria-hidden="true" />
                <span>{item.label}</span>
                {item.badge && <b>{item.badge}</b>}
              </button>
            ))}
          </nav>
        </aside>

        <section className="account-content">
          <div className="account-title-row">
            <div>
              <span>Личный кабинет</span>
              <h1>Добрый вечер, {activeProfile.name}</h1>
            </div>
            <button type="button" className="account-title-row__settings">
              <i className="hgi hgi-stroke hgi-settings-02" aria-hidden="true" />
              Настройки
            </button>
          </div>

          <section className="account-year" aria-labelledby="account-year-title">
            <div className="account-year__copy">
              <span className="account-year__eyebrow">Твоя история · 2025</span>
              <h2 id="account-year-title">Вот к чему ты возвращался весь год</h2>
              <p>
                Вспомни самые заметные моменты и посмотри, чем тебе запомнился этот год на Авито.
              </p>

              <div className="account-year__clue">
                <span className="account-year__clue-icon" aria-hidden="true">
                  <i className="hgi hgi-stroke hgi-favourite" />
                </span>
                <div>
                  <span>Кое-что уже готово</span>
                  <strong>Посмотри, что стало твоей главной темой года</strong>
                </div>
              </div>

              <div className="account-year__actions">
                <Link to="/year" className="account-year__cta">
                  Что там у меня?
                  <i className="hgi hgi-stroke hgi-arrow-right-01" aria-hidden="true" />
                </Link>
              </div>
            </div>

            <div className="account-year__preview" aria-hidden="true">
              <div className="account-year__preview-head">
                <span>Мой 2025</span>
                <b>Авито</b>
              </div>
              <div className="account-year__number">72</div>
              <div className="account-year__number-copy">дня с Авито</div>
              <div className="account-year__hidden">
                <small>Главная тема</small>
                <strong>██████████</strong>
              </div>
              <div className="account-year__hidden account-year__hidden--second">
                <small>Тип года</small>
                <strong>██████████████</strong>
              </div>
              <span className="account-year__orb account-year__orb--blue" />
              <span className="account-year__orb account-year__orb--green" />
            </div>
          </section>

          <section className="account-overview" aria-label="Обзор аккаунта">
            <article className="account-overview-card">
              <div className="account-overview-card__topline">
                <span className="account-overview-card__icon"><i className="hgi hgi-stroke hgi-wallet-02" aria-hidden="true" /></span>
                <button type="button">Пополнить</button>
              </div>
              <span>Кошелёк</span>
              <strong>1 280 ₽</strong>
            </article>

            <article className="account-overview-card">
              <div className="account-overview-card__topline">
                <span className="account-overview-card__icon"><i className="hgi hgi-stroke hgi-favourite" aria-hidden="true" /></span>
                <span className="account-overview-card__badge">+3</span>
              </div>
              <span>Избранное</span>
              <strong>18 объявлений</strong>
              <p>Появились новые предложения</p>
            </article>

            <article className="account-overview-card">
              <div className="account-overview-card__topline">
                <span className="account-overview-card__icon"><i className="hgi hgi-stroke hgi-message-01" aria-hidden="true" /></span>
                <span className="account-overview-card__badge account-overview-card__badge--red">3</span>
              </div>
              <span>Сообщения</span>
              <strong>Есть непрочитанные</strong>
              <p>Последнее — 12 минут назад</p>
            </article>
          </section>

          <section className="account-listings">
            <div className="account-listings__heading">
              <div>
                <span>Мои объявления</span>
                <h2>Активные</h2>
              </div>
              <button type="button">Все объявления</button>
            </div>

            <div className="account-listing-row">
              <div className="account-listing-row__image" aria-hidden="true" />
              <div className="account-listing-row__copy">
                <strong>Комплект мебели для гостиной</strong>
                <span>24 000 ₽ · опубликовано 5 дней назад</span>
              </div>
              <span className="account-listing-row__status">Активно</span>
              <button type="button" aria-label="Открыть объявление">
                <i className="hgi hgi-stroke hgi-arrow-right-01" aria-hidden="true" />
              </button>
            </div>
          </section>
        </section>
      </main>
    </div>
  );
}
