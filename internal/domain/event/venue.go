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
	Events      []Event
}

func AllVenues() []*Venue {
	return []*Venue{
		{
			ID:          VenueIDYokohamaArena,
			DisplayName: "横浜アリーナ",
			Emoji:       "🏟️",
			Events:      []Event{},
		},
		{
			ID:          VenueIDNissanStadium,
			DisplayName: "日産スタジアム",
			Emoji:       "⚽",
			Events:      []Event{},
		},
		{
			ID:          VenueIDSkateCenter,
			DisplayName: "KOSÉ新横浜スケートセンター",
			Emoji:       "⛸️",
			Events:      []Event{},
		},
	}
}

func VenueByID(id VenueID) *Venue {
	for _, v := range AllVenues() {
		if v.ID == id {
			return v
		}
	}
	return nil
}
