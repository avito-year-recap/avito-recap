package recap

import "time"

type UserID uint64

// EventType перечисляет действия пользователя, которые recap считает
// значимым сигналом. Держать в синхроне со значениями
// `events.event_type` LowCardinality в ClickHouse (clickhouse/init/001_schema.sql).
type EventType string

const (
	EventView         EventType = "view_ad"
	EventSearch       EventType = "search"
	EventFavoriteAdd  EventType = "favorite_add"
	EventMessageSent  EventType = "message_sent"
	EventCallMade     EventType = "call_made"
	EventPriceOffer   EventType = "price_offer"
	EventAdCreated    EventType = "ad_created"
	EventAdPublished  EventType = "ad_published"
	EventAdSold       EventType = "ad_sold"
	EventPurchaseMade EventType = "purchase_made"
	EventReviewLeft   EventType = "review_left"
)

// AdStatusEvents — подмножество EventType, отражающее смену статуса
// объявления (created → published → sold). Порядок важен: используется там,
// где по последнему событию для ad_id определяется текущий статус
// (черновик/активно/продано).
var AdStatusEvents = []EventType{EventAdCreated, EventAdPublished, EventAdSold}

// Event is a single tracked user action — the raw fact written to
// ClickHouse that every recap card is ultimately derived from.
type Event struct {
	ID          string // optional; a UUID is generated on insert if empty
	UserID      UserID
	Type        EventType
	OccurredAt  time.Time
	Category    string
	Subcategory string
	City        string
	Price       *float64
	AdID        *uint64 // объявление: view_ad/ad_created/ad_published/ad_sold
	DialogID    *uint64 // диалог: message_sent (начатый чат) и purchase_made (чат, приведший к покупке)
	SearchQuery *string
	Metadata    map[string]string
}

// Visibility decides whether a card is safe to include on a public
// share-card. Public cards may only carry rounded/derived storytelling data
// (totals, top category, persona, badges) — never raw search text, exact
// prices, or anything that implies contact with another user.
type Visibility string

const (
	VisibilityPrivate Visibility = "private" // shown only in the user's own recap
	VisibilityPublic  Visibility = "public"  // safe to render on a shared card
)

// CardType is a fixed taxonomy of 8 slide kinds. Every recap is assembled
// from exactly these types (0..N instances each — most appear once, but
// "achievement" can repeat).
type CardType string

const (
	CardIntro         CardType = "intro"          // cover slide
	CardTotals        CardType = "totals"          // year-in-numbers headline
	CardTopCategory   CardType = "top_category"     // the category the user is most invested in
	CardPersona       CardType = "persona"          // the one archetype assigned for the year
	CardRhythm        CardType = "rhythm"           // when: night owl / favorite weekday / peak month
	CardSearchInsight CardType = "search_insight"   // what the user searched for most
	CardAchievement   CardType = "achievement"      // one badge earned (repeatable)
	CardNextStep      CardType = "next_step"        // closing slide: missed-opportunity note + CTA
)

// Metric is the single headline number a card is built around, e.g.
// "128 действий" or "70% в категории Транспорт".
type Metric struct {
	Value float64 `json:"value"`
	Unit  string  `json:"unit,omitempty"`
	Label string  `json:"label"`
}

// NextAction is the concrete next step offered after (or on) a card — the
// product's answer to "what do I do with this now".
type NextAction struct {
	Type      string `json:"type"` // return_to_category | reopen_favorites | continue_search | create_ad
	Label     string `json:"label"`
	TargetURL string `json:"target_url"`
}

// Card is one slide of the recap.
type Card struct {
	ID         string      `json:"id"`
	Type       CardType    `json:"type"`
	Order      int         `json:"order"`
	Visibility Visibility  `json:"visibility"`
	Title      string      `json:"title"`
	Subtitle   string      `json:"subtitle,omitempty"`
	Emoji      string      `json:"emoji,omitempty"`
	Metric     *Metric     `json:"metric,omitempty"`
	Reason     string      `json:"reason,omitempty"` // why the user got this card, shown in-app
	CTA        *NextAction `json:"cta,omitempty"`
}

// Persona is the single archetype assigned to the user for the year — the
// through-line that ties the individual cards into one story.
type Persona struct {
	Code        string `json:"code"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Reason      string `json:"reason"`
}

// Recap is the full generated "Итоги года" result for one user/year.
// Persona and NextAction are duplicated onto the top level (in addition to
// appearing as a persona/next_step Card) because the frontend needs them
// outside the swipeable deck too — e.g. a page title or a persistent CTA
// button.
type Recap struct {
	UserID      UserID     `json:"user_id"`
	Year        int        `json:"year"`
	GeneratedAt time.Time  `json:"generated_at"`
	Persona     Persona    `json:"persona"`
	Cards       []Card     `json:"cards"`
	NextAction  NextAction `json:"next_action"`
}

// PublicCards returns the subset of cards safe to render on an
// outward-facing share-card (e.g. shared to a chat or social network).
func (r Recap) PublicCards() []Card {
	return r.filterCards(func(c Card) bool { return c.Visibility == VisibilityPublic })
}

// CardsByType returns the cards of one type, in deck order. Mainly useful
// for "achievement", the one repeatable type.
func (r Recap) CardsByType(t CardType) []Card {
	return r.filterCards(func(c Card) bool { return c.Type == t })
}

func (r Recap) filterCards(keep func(Card) bool) []Card {
	out := make([]Card, 0, len(r.Cards))
	for _, c := range r.Cards {
		if keep(c) {
			out = append(out, c)
		}
	}
	return out
}
