import { Link } from "react-router-dom";
import { resolveProfileAvatarUrl } from "../../entities/profile-avatar";
import { getActiveProfile } from "../../shared/lib/active-profile";
import { AppLogo } from "../../shared/ui/AppLogo";
import "./AvitoHomePage.css";

const categories = [
  { icon: "hgi-car-01", label: "Авто", tone: "blue" },
  { icon: "hgi-chair-01", label: "Для дома", tone: "green" },
  { icon: "hgi-smart-phone-01", label: "Электроника", tone: "purple" },
  { icon: "hgi-bicycle-01", label: "Хобби", tone: "coral" },
];

const listings = [
  { title: "Кресло из массива", price: "8 500 ₽", meta: "Сегодня · рядом", tone: "chair" },
  { title: "Смартфон, 256 ГБ", price: "42 000 ₽", meta: "1 час назад · доставка", tone: "phone" },
  { title: "Городской велосипед", price: "18 900 ₽", meta: "Вчера · доставка", tone: "bike" },
  { title: "Рабочий стол", price: "12 500 ₽", meta: "Сегодня · Москва", tone: "table" },
];

export function AvitoHomePage() {
  const activeProfile = getActiveProfile();
  const activeAvatarUrl = resolveProfileAvatarUrl(activeProfile.profileCode, activeProfile.avatarUrl);

  return (
    <div className="avito-home">
      <header className="avito-home__header">
        <div className="avito-home__header-inner">
          <Link to="/" className="avito-home__brand" aria-label="Авито">
            <AppLogo compact />
          </Link>

          <nav className="avito-home__nav" aria-label="Основная навигация">
            <button type="button" className="avito-home__nav-item">
              <i className="hgi hgi-stroke hgi-favourite" aria-hidden="true" />
              <span>Избранное</span>
            </button>
            <button type="button" className="avito-home__nav-item avito-home__nav-item--desktop">
              <i className="hgi hgi-stroke hgi-message-01" aria-hidden="true" />
              <span>Сообщения</span>
              <b aria-label="3 непрочитанных">3</b>
            </button>
            <Link to="/account" className="avito-home__cabinet-link" aria-label="Открыть личный кабинет">
              <i className="hgi hgi-stroke hgi-user" aria-hidden="true" />
              <span>Личный кабинет</span>
            </Link>
            <Link
              to="/profiles?return=/"
              className="avito-home__account-control"
              aria-label={`Сменить аккаунт. Сейчас ${activeProfile.name}`}
              title="Сменить аккаунт"
            >
              <span className="avito-home__avatar-link" aria-hidden="true">
                <img className="avito-home__avatar" src={activeAvatarUrl} alt="" />
                <span className="avito-home__switch-dot">
                  <i className="hgi hgi-stroke hgi-arrow-right-01" />
                </span>
              </span>
              <span className="avito-home__account-name">{activeProfile.name}</span>
            </Link>
            <button type="button" className="avito-home__post">Разместить объявление</button>
          </nav>
        </div>
      </header>

      <main className="avito-home__main">
        <section className="avito-search" aria-label="Поиск объявлений">
          <button type="button" className="avito-search__category">Все категории</button>
          <label className="avito-search__field">
            <span className="sr-only">Поиск по объявлениям</span>
            <input type="search" placeholder="Найти на Авито" />
          </label>
          <button type="button" className="avito-search__submit">
            <i className="hgi hgi-stroke hgi-search-01" aria-hidden="true" />
            <span>Найти</span>
          </button>
        </section>

        <button type="button" className="avito-location">
          <i className="hgi hgi-stroke hgi-location-01" aria-hidden="true" />
          Москва
        </button>

        <section className="avito-categories" aria-label="Популярные категории">
          {categories.map((category) => (
            <button key={category.label} type="button" className="avito-category">
              <span className={`avito-category__icon avito-category__icon--${category.tone}`} aria-hidden="true">
                <i className={`hgi hgi-stroke ${category.icon}`} />
              </span>
              <span>{category.label}</span>
            </button>
          ))}
        </section>

        <section className="recap-entry" aria-labelledby="recap-entry-title">
          <div className="recap-entry__copy">
            <span className="recap-entry__eyebrow">
              <span className="recap-entry__live-dot" aria-hidden="true" />
              {activeProfile.name}, твой 2025 на Авито
            </span>

            <h1 id="recap-entry-title">Кажется, в этом году ты кое-чем увлёкся</h1>
            <p>
              Вспомни, что искал, сохранял и покупал — и узнай, чем запомнился твой год на Авито.
            </p>

            <div className="recap-entry__proof" aria-label="Один факт из итогов года">
              <span className="recap-entry__proof-icon" aria-hidden="true">
                <i className="hgi hgi-stroke hgi-favourite" />
              </span>
              <div>
                <strong>Есть кое-что, к чему ты возвращался чаще всего</strong>
                <span>Посмотри, что стало твоей темой года.</span>
              </div>
            </div>

            <div className="recap-entry__actions">
              <Link to="/year" className="recap-entry__cta">
                Открыть мой 2025
                <i className="hgi hgi-stroke hgi-arrow-right-01" aria-hidden="true" />
              </Link>
            </div>
          </div>

          <div className="recap-entry__visual" aria-hidden="true">
            <div className="recap-entry__poster">
              <div className="recap-entry__poster-topline">
                <span>Мой год на Авито</span>
                <span>2025</span>
              </div>
              <strong className="recap-entry__poster-year">20<br />25</strong>

              <div className="recap-entry__secret recap-entry__secret--top">
                <small>Твоя тема года</small>
                <b className="recap-entry__blur-text">Электроника</b>
              </div>
              <div className="recap-entry__secret recap-entry__secret--bottom">
                <small>Твой тип года</small>
                <b className="recap-entry__blur-text">Охотник за находками</b>
              </div>

              <span className="recap-entry__shape recap-entry__shape--blue" />
              <span className="recap-entry__shape recap-entry__shape--green" />
              <span className="recap-entry__shape recap-entry__shape--purple" />
            </div>
          </div>
        </section>

        <section className="avito-feed" aria-labelledby="avito-feed-title">
          <div className="avito-feed__heading">
            <h2 id="avito-feed-title">Рекомендации для вас</h2>
            <button type="button">Показать ещё</button>
          </div>

          <div className="avito-feed__grid">
            {listings.map((listing) => (
              <article key={listing.title} className="avito-listing">
                <div className={`avito-listing__image avito-listing__image--${listing.tone}`}>
                  <button type="button" className="avito-listing__favorite" aria-label={`Добавить «${listing.title}» в избранное`}>
                    <i className="hgi hgi-stroke hgi-favourite" aria-hidden="true" />
                  </button>
                  <span className="avito-listing__object" aria-hidden="true" />
                </div>
                <div className="avito-listing__copy">
                  <strong>{listing.price}</strong>
                  <h3>{listing.title}</h3>
                  <p>{listing.meta}</p>
                </div>
              </article>
            ))}
          </div>
        </section>
      </main>
    </div>
  );
}
