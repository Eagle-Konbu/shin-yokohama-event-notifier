package event

type VenueID string

const (
	VenueIDYokohamaArena VenueID = "yokohama_arena"
	VenueIDNissanStadium VenueID = "nissan_stadium"
	VenueIDSkateCenter   VenueID = "skate_center"
)

type Venue struct {
	ID          VenueID
	DisplayName string
	Emoji       string
}

var (
	YokohamaArena = Venue{
		ID:          VenueIDYokohamaArena,
		DisplayName: "横浜アリーナ",
		Emoji:       "🏟️",
	}
	NissanStadium = Venue{
		ID:          VenueIDNissanStadium,
		DisplayName: "日産スタジアム",
		Emoji:       "⚽",
	}
	SkateCenter = Venue{
		ID:          VenueIDSkateCenter,
		DisplayName: "KOSÉ新横浜スケートセンター",
		Emoji:       "⛸️",
	}
)

func AllVenues() []Venue {
	return []Venue{
		YokohamaArena,
		NissanStadium,
		SkateCenter,
	}
}

func VenueByID(id VenueID) Venue {
	for _, v := range AllVenues() {
		if v.ID == id {
			return v
		}
	}
	return Venue{}
}
