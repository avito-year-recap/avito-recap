package cards

func MonthName(month uint32) string {
	names := [...]string{"", "январь", "февраль", "март", "апрель", "май", "июнь", "июль", "август", "сентябрь", "октябрь", "ноябрь", "декабрь"}
	if month >= uint32(len(names)) {
		return ""
	}
	return names[month]
}
