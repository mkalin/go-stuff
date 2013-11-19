package main

import (
	"fmt"
	"time"
)

// An Event stands in temporal relationships with other Events and has a 
// positive Duration. 
type Event struct {
	Id            int
	Name          string
	Description   string
	Start         time.Time
	Finish        time.Time
	Duration      time.Duration
	DurationScale time.Duration  // hour,...,second,...,nanosecond (default)
	TempRelations []*Relation
	Incoming      []*Event       // for topological sorting
	IncomingN     int            // ditto
	Outgoing      []*Event       // ditto
}

// Enable sorting.
type Events []*Event

// ByStart implements sort.Interface by providing Less and using the Len and
// Swap methods of the embedded Organs value.
type ByStart struct { Events }

func (s Events) Len() int { return len(s) }
func (s Events) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s ByStart) Less(i, j int) bool { return s.Events[i].Start.Before(s.Events[j].Start) }

//** Temporal relations implemented as 2-arg functions
type TempFunc func(*Event, *Event) bool
type Relation struct {
	tempFunc TempFunc
	event    *Event
}

//** sample temporal relations: extend as desired
// Some relations are definable via others. For example, StartAfterStart
// could be defined as 
//      !(StartBeforeStart || StartAtStart)
// The relations listed here are simply relatively basic examples.
func FinishAfterFinish(e1, e2 *Event) bool {	return e1.Finish.After(e2.Finish) }
func FinishAfterStart(e1, e2 *Event) bool { return e1.Finish.After(e2.Start) }
func FinishBeforeStart(e1, e2 *Event) bool { return e1.Finish.Before(e2.Start) }
func FinishBeforeFinish(e1, e2 *Event) bool { return e1.Finish.Before(e2.Finish) }
func FinishAtFinish(e1, e2 *Event) bool {	return e1.Finish.Equal(e2.Finish) }
func FinishAtStart(e1, e2 *Event) bool { return e1.Finish.Equal(e2.Start) }

func StartAfterStart(e1, e2 *Event) bool { return e1.Start.After(e2.Start) }
func StartAfterFinish(e1, e2 *Event) bool { return e1.Start.After(e2.Finish) }
func StartBeforeStart(e1, e2 *Event) bool { return e1.Start.Before(e2.Start) }
func StartBeforeFinish(e1, e2 *Event) bool { return e1.Start.Before(e2.Finish) }
func StartAtStart(e1, e2 *Event) bool { return e1.Start.Equal(e2.Start) }
func StartAtFinish(e1, e2 *Event) bool { return e1.Start.Equal(e2.Finish) }

func During(e1, e2 *Event) bool { return e1.Start.After(e2.Start) && e1.Finish.Before(e2.Finish) }

//** methods
func (e *Event) String() string {
	froms := []int{}
	tos := []int{}

	for _, from := range e.Incoming {
		froms = append(froms, from.Id)
	}
	for _, to := range e.Outgoing {
		tos = append(tos, to.Id)
	}

	return fmt.Sprintf("%s:\n\tId:\t%1d\n\tStart:\t%v (%v)\n\tFrom:\t%v (%v in all)\n\tTo:\t%v (%v in all)\n", 
		e.Name, e.Id, e.Start, e.Start.UnixNano(), froms, len(froms), tos, len(tos))
}

//** topological sort support
func computeOutgoing(hash map[int]*Event) {
	for _, event := range hash {
		for _, from := range event.Incoming {
			from.Outgoing = append(from.Outgoing, event)
		}
	}
}

func setIncomingCounts(hash map[int]*Event) {
	for _, event := range hash {
		event.IncomingN = len(event.Incoming)
	}
}

func listifyMap(hash map[int]*Event) []*Event {
	list := []*Event{}

	for _, event := range hash {
		list = append(list, event)
	}
	return list
}

func topSort(eventMap map[int]*Event, list []*Event) []*Event {
	setIncomingCounts(eventMap)
	sorted := []*Event{}
	nopreds := []*Event{}

	// Initial list consists of nodes with no incoming edges.
	for _, event := range list {
		if event.IncomingN == 0 {
			nopreds = append(nopreds, event)
		}
	}

	// Sort.
	for len(nopreds) > 0 {
		// Pick an event from the nopreds list and add it to the sorted list.
		event := nopreds[0]
		sorted = append(sorted, event)
		
		// Remove the picked event from the nopreds list.
		nopreds[len(nopreds) - 1], nopreds[0], nopreds = 
			nil, nopreds[len(nopreds) - 1], nopreds[:len(nopreds) - 1]
		
		for _, to := range event.Outgoing {
			to.IncomingN--                    // "remove" the edge
			if to.IncomingN == 0 {            // any more incoming edges?
				nopreds = append(nopreds, to)  // if not, add node to nopreds list
			}
		}
	}
	for _, event := range list {
		if event.IncomingN > 0 {
			return nopreds        // cycle detected: return empty list
		}
	}
	return sorted               // non-empty, topologically sorted list
}

//** utilities
func dumpList(msg string, list []*Event) {
	fmt.Println(msg)
	for _, value := range list {
		fmt.Println(value.String())
	}
}

func addConstraint(e1 *Event, e2 *Event, f TempFunc) {
	r := &Relation { tempFunc: f, event: e2 }
	e1.TempRelations = append(e1.TempRelations, r)
	e1.Incoming = append(e1.Incoming, e2)
}

func buildConstraints(hash map[int]*Event) {
	//  Fixing the car (1) is During planning (0)
	e1 := hash[1]
	e2 := hash[0]
	addConstraint(e1, e2, During)

	// Preparing the luggage (2) is During planning (0)
	e1 = hash[2]
	e2 = hash[0]
	addConstraint(e1, e2, During)

	// Loading the luggage (3) is After preparing the luggage (2)
	e1 = hash[3]
	e2 = hash[2]
	addConstraint(e1, e2, FinishBeforeStart)

	// Fixing the car (1) is Before gassing up the car (4)
	e1 = hash[4]
	e2 = hash[1]
	addConstraint(e1, e2, FinishBeforeStart)

	// Final check (5) is AtFinish of gassing up the car (4)
	e1 = hash[5]
	e2 = hash[4]
	addConstraint(e1, e2, StartAtFinish)

	// Start driving (6) After final check (5) and luggage loaded (3)
	e1 = hash[6]
	e2 = hash[5]
	addConstraint(e1, e2, StartAfterFinish)
	e2 = hash[3]
	addConstraint(e1, e2, StartAfterFinish)

	// Drive to destination (7) After starting (6)
	e1 = hash[7]
	e2 = hash[6]
	addConstraint(e1, e2, StartAfterFinish)

	// Eat (8) while driving (7).
	e1 = hash[8]
	e2 = hash[7]
	addConstraint(e1, e2, During)

	// Unload (9) After driving to destination (7)
	e1 = hash[9]
	e2 = hash[7]
	addConstraint(e1, e2, StartAtFinish)
}

func buildEvent(eventMap map[int]*Event, id int, name string, desc string, dur time.Duration) {
	event :=  new(Event)
	event.Id = id
	event.Name = name
	event.Description = desc
	event.DurationScale = time.Minute
	event.Duration = time.Duration(dur) * event.DurationScale
	eventMap[id] = event
}

func buildEvents() map[int]*Event {
	eventMap := make(map[int]*Event)

	buildEvent(eventMap, 0, "Plan trip", "Plan car trip", 201)
	buildEvent(eventMap, 1, "Fix junker", "Repair the car as needed", 505)
	buildEvent(eventMap, 2, "Prepare luggage", "Begin to pack up the stuff (but not the kids and dog)", 317)
	buildEvent(eventMap, 3, "Load luggage", "Load the luggage (inclduing the kids and dog)", 127)
	buildEvent(eventMap, 4, "Gas up car", "Fill the gas tank", 12)
	buildEvent(eventMap, 5, "Final check", "Check that all's ready", 1)
	buildEvent(eventMap, 6, "Start trip", "Start driving", 1)
	buildEvent(eventMap, 7, "Drive", "Drive to destination", 819)
	buildEvent(eventMap, 8, "Eat and drive", "Eat while driving", 18)
	buildEvent(eventMap, 9, "Unload", "Unload the luggage", 42)

	return eventMap
}

func buildExample() map[int]*Event {
	eventMap := buildEvents()
	buildConstraints(eventMap)
	computeOutgoing(eventMap)
	return eventMap
}

func main() {
	eventMap := buildExample()
	list := listifyMap(eventMap)
	dumpList("Original list:\n", list)
	//sortedList := topSort(eventMap, list)
	//dumpList("Sorted list:\n", sortedList)
}