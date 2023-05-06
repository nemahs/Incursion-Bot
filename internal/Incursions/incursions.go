package incursions

import (
	"IncursionBot/internal/ESI"
	logging "IncursionBot/internal/Logging"
	"fmt"
	"strings"
	"time"
)

const mobilizingLifetime time.Duration = time.Hour * 72
const withdrawingLifetime time.Duration = time.Hour * 24
const establishedMaxLife time.Duration = time.Hour * 24 * 8

type IncursionState string

const (
	Established IncursionState = "established"
	Mobilizing  IncursionState = "mobilizing"
	Withdrawing IncursionState = "withdrawing"
	Respawning  IncursionState = "respawning"
	Unknown     IncursionState = ""
)

func ParseState(val string) IncursionState {
	toTest := strings.ToLower(val)

	switch toTest {
	case string(Established):
		return Established
	case string(Mobilizing):
		return Mobilizing
	case string(Withdrawing):
		return Withdrawing
	default:
		return Unknown
	}
}

type NamedItem struct {
	Name string
	ID   int
}

func CreateNamedItem(id int, esi *ESI.ESIClient) (item NamedItem) {
	item.ID = id
	item.Name = "Unknown" // Default value

	name, err := esi.GetNames([]int{id})

	if err != nil {
		logging.Errorf("Error occurred getting name for %d: %v", id, err)
		return
	}

	item.Name = name[id]
	return
}

type Incursion struct {
	Constellation NamedItem       // Constellation the incursion is in
	Layout        IncursionLayout // Layout of the spawn
	SovOwner      string          // TCU owner of the staging system
	Influence     float64         // Influence of the incursion from 0 to 1 inclusive
	Region        NamedItem       // Region the incursion is in
	State         IncursionState  // Current state of the incursion
	Security      SecurityClass   // Security type of the staging system
	SecStatus     float64         // Security status of the staging system, -1 to 1 inclusive
	Distance      int             // Distance from home system
	StateChanged  time.Time       // Time the state changed to this current state
	Systems       []int           // IDs for all systems in the spawn
	IsValid       bool
}

func (inc *Incursion) Equal(other Incursion) bool {
	return inc.Layout.StagingSystem.ID == other.Layout.StagingSystem.ID
}

func (inc *Incursion) ToString() string {
	var sovString string

	if inc.SovOwner != "" {
		sovString = fmt.Sprintf("[%s] ", inc.SovOwner)
	}

	return fmt.Sprintf("%s %s{%.2f} (HQ: %s) (%s - %s)",
		inc.Layout.StagingSystem.Name,
		sovString,
		inc.SecStatus,
		inc.Layout.HQSystem.Name,
		inc.Constellation.Name,
		inc.Region.Name)
}

func (inc *Incursion) TimeLeftInSpawn() (time.Time, error) {
	logging.Infof("Stage changed: %s", inc.StateChanged)

	switch inc.State {
	case Established:
		return inc.StateChanged.Add(establishedMaxLife), nil
	case Mobilizing:
		return inc.StateChanged.Add(mobilizingLifetime), nil
	case Withdrawing:
		return inc.StateChanged.Add(withdrawingLifetime), nil
	}

	return time.Time{}, fmt.Errorf("not a state we can deal with")
}

func (inc *Incursion) TimeLeftString(timeFormat string) string {
	if inc.StateChanged.IsZero() {
		return "Unknown"
	}

	despawn, err := inc.TimeLeftInSpawn()
	if err != nil {
		logging.Errorf("Error occurred getting time left in spawn %s: %s", inc.Layout.StagingSystem.Name, err)
		return "Unknown"
	}

	logging.Debugf("Despawn result: %s", despawn)
	if inc.State == Established {
		return fmt.Sprintf("NLT %s", despawn.UTC().Format(timeFormat))
	}

	return fmt.Sprintln(despawn.UTC().Format(timeFormat))
}

// Updates the give incursion wih new data. Returns true if the state changed, False otherwise.
func (incursion *Incursion) Update(influence float64, state IncursionState) bool {
	incursion.Influence = influence

	if incursion.State != state {
		incursion.State = state
		incursion.StateChanged = time.Now()
		return true
	}

	return false
}
