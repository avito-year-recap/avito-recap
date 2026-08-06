package analytics

// CategoryDefinition is the product catalogue subset understood by recap.
// Analytics may still return other safe category codes; they simply will not
// produce a thematic achievement until explicitly mapped here.
type CategoryDefinition struct {
	Code      string `json:"code"`
	Title     string `json:"title"`
	Shareable bool   `json:"shareable"`
}

const (
	CategoryElectronics     = "electronics"
	CategoryAuto            = "auto"
	CategoryRealEstate      = "real_estate"
	CategoryFurniture       = "furniture"
	CategoryClothing        = "clothing"
	CategoryHobby           = "hobby"
	CategorySports          = "sports"
	CategoryServices        = "services"
	CategoryWomensFashion   = "womens_fashion"
	CategoryBeautyCosmetics = "beauty_cosmetics"
	CategoryMensFashion     = "mens_fashion"
	CategoryOutdoorTravel   = "outdoor_travel"
	CategoryGarden          = "garden"
	CategoryBooks           = "books"
	CategoryJewelry         = "jewelry"
	CategoryMusic           = "music"
	CategoryToysDolls       = "toys_dolls"
	CategoryTools           = "tools"
	CategoryPets            = "pets"
	CategoryKids            = "kids"
)

var categoryCatalogue = []CategoryDefinition{
	{Code: CategoryElectronics, Title: "Электроника", Shareable: true},
	{Code: CategoryAuto, Title: "Авто", Shareable: true},
	{Code: CategoryRealEstate, Title: "Недвижимость", Shareable: true},
	{Code: CategoryFurniture, Title: "Мебель и интерьер", Shareable: true},
	{Code: CategoryClothing, Title: "Одежда, обувь и аксессуары", Shareable: true},
	{Code: CategoryHobby, Title: "Хобби и отдых", Shareable: true},
	{Code: CategorySports, Title: "Спорт и отдых", Shareable: true},
	{Code: CategoryServices, Title: "Услуги", Shareable: true},
	{Code: CategoryWomensFashion, Title: "Женская одежда и аксессуары", Shareable: true},
	{Code: CategoryBeautyCosmetics, Title: "Красота и косметика", Shareable: false},
	{Code: CategoryMensFashion, Title: "Мужская одежда и аксессуары", Shareable: true},
	{Code: CategoryOutdoorTravel, Title: "Туризм и путешествия", Shareable: true},
	{Code: CategoryGarden, Title: "Дача и сад", Shareable: true},
	{Code: CategoryBooks, Title: "Книги", Shareable: true},
	{Code: CategoryJewelry, Title: "Украшения и ювелирные изделия", Shareable: false},
	{Code: CategoryMusic, Title: "Музыкальные инструменты и аудио", Shareable: true},
	{Code: CategoryToysDolls, Title: "Игрушки, куклы и коллекционирование", Shareable: true},
	{Code: CategoryTools, Title: "Инструменты", Shareable: true},
	{Code: CategoryPets, Title: "Товары для животных", Shareable: false},
	{Code: CategoryKids, Title: "Детские товары", Shareable: false},
}

// CategoryCatalogue returns a defensive copy suitable for API projection or
// analytics configuration.
func CategoryCatalogue() []CategoryDefinition {
	return append([]CategoryDefinition(nil), categoryCatalogue...)
}
