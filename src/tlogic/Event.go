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

var eventMap map[int]*Event

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
	return ""
}

//** topological sort support
func computeOutgoing(list []*Event) {
	for _, event := range list {
		for _, from := range event.Incoming {
			i := from.Id
			list[i].Outgoing = append(list[i].Outgoing, event)
		}
	}
}

func setIncomingCounts(list []*Event) {
	for _, event := range list {
		event.IncomingN = len(event.Incoming)
	}
}

func topSort(list []*Event) []*Event {
	setIncomingCounts(list)
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
func dumpMap(msg string, hash map[int]*Event) {
	fmt.Println(msg)
	for _, value := range hash {
		fmt.Println(value.String())
	}
}

func buildConstraints() {
	// e1 (plan) is During e2 (fix car)
	e1 := eventMap[1]
	e2 := eventMap[0]
	tr := &Relation {
		tempFunc: During,
	   event:    e2}
	e1.TempRelations = append(e1.TempRelations, tr)
	e1.Incoming = append(e1.Incoming, e2)

	// e1 (plan) is also During e2 (prepare luggage)
	e2 = eventMap[2]
	tr = &Relation {
		tempFunc: During,
	   event:    e2}
	e1.TempRelations = append(e1.TempRelations, tr)
	e1.Incoming = append(e1.Incoming, e2)

	// e2 (load luggage) is After e1 (prepare luggage)
	e1 = eventMap[3]
	e2 = eventMap[2]
	tr = &Relation {
		tempFunc: FinishBeforeStart,
	   event:    e2}
	e1.TempRelations = append(e1.TempRelations, tr)
	e2.Incoming = append(e2.Incoming, e1)
	
	// e1 (load luggage) is Before e2 (gas up car)
	e1 = eventMap[3]
	e2 = eventMap[4]	
	tr = &Relation {
		tempFunc: FinishBeforeStart,
	   event:    e2}
	e1.TempRelations = append(e1.TempRelations, tr)
	e2.Incoming = append(e2.Incoming, e1)

	// e2 (start car) is AtFinish of e1 (gas up car)
	e1 = eventMap[4]
	e2 = eventMap[5]
	tr = &Relation {
		tempFunc: StartAtFinish,
	   event:       e1}
	e2.TempRelations = append(e2.TempRelations, tr)
	e2.Incoming = append(e2.Incoming, e1)

	// commence driving (e2) when car is started (e1)
	e1 = eventMap[5]
	e2 = eventMap[6]
	tr = &Relation {
		tempFunc: StartAtFinish,
	   event:       e1}
	e2.TempRelations = append(e2.TempRelations, tr)
	e2.Incoming = append(e2.Incoming, e1)

	// drive to destination (e2) once driving has begun (e1)
	e1 = eventMap[6]
	e2 = eventMap[7]
	tr = &Relation {
		tempFunc: StartAtFinish,
	   event:       e1}
	e2.TempRelations = append(e2.TempRelations, tr)
	e2.Incoming = append(e2.Incoming, e1)

	// stop at destination (e2) once the driving there is done (e1)
	e1 = eventMap[7]
	e2 = eventMap[8]
	tr = &Relation {
		tempFunc: StartAtFinish,
	   event:       e1}
	e2.TempRelations = append(e2.TempRelations, tr)
	e2.Incoming = append(e2.Incoming, e1)
	
	// unpack (e2) once the driving is over (e1)
	e1 = eventMap[8]
	e2 = eventMap[9]
	tr = &Relation {
		tempFunc: StartAtFinish,
	   event:       e1}
	e2.TempRelations = append(e2.TempRelations, tr)
	e2.Incoming = append(e2.Incoming, e1)
}

func buildEvents() {
	/* Sample problem with sample events:

	 Problem: Take an automobile trip from X to Y.
	 
	 Events:
	 -- service the automobile (0)
	 -- plan the trip          (1)
	 -- prepare the luggage    (2)
	 -- load the luggage       (3)
	 -- gas up the car         (4)            
	 -- start the car          (5)
	 -- commence driving       (6)
	 -- drive to destination   (7)
	 -- stop the car           (8)
	 -- unload the luggage     (9)
	 */
	eventMap = make(map[int]*Event)
	
	// sample model
	event := new(Event)
	event.Id = 0
	event.Name = "Event-plan"
	event.Description = "Plan the car trip"
	event.DurationScale = time.Minute
	event.Duration = time.Duration(201) * event.DurationScale
	eventMap[0] = event
	
	event = new(Event)
	event.Id = 1
	event.Name = "Event-FixJunker"
	event.Description = "Repair the car as needed"
	event.DurationScale = time.Minute
	event.Duration = time.Duration(505) * event.DurationScale
	eventMap[1] = event

	event = new(Event)
	event.Id = 2
	event.Name = "Event-Pack"
	event.Description = "Pack up the stuff (but not the kIds and dog)"
	event.DurationScale = time.Minute
	event.Duration = time.Duration(317) * event.DurationScale 
	eventMap[2] = event

	event = new(Event)
	event.Id = 3
	event.Name = "Event-Load"
	event.Description = "Load the luggage (inclduing the kIds and dog)"
	event.DurationScale = time.Minute
	event.Duration = time.Duration(127) * event.DurationScale
	eventMap[3] = event

	event = new(Event)
	event.Id = 4
	event.Name = "Event-GasUp"
	event.Description = "Fill the gas tank"
	event.DurationScale = time.Minute
	event.Duration = time.Duration(12) * event.DurationScale
	eventMap[4] = event

	event = new(Event)
	event.Id = 5
	event.Name = "Event-StartCar"
	event.Description = "Crank up the junker"
	event.DurationScale = time.Minute
	event.Duration = time.Duration(1) * event.DurationScale
	eventMap[5] = event

	event = new(Event)
	event.Id = 6
	event.Name = "Event-Commence"
	event.Description = "Start driving"
	event.DurationScale = time.Minute
	event.Duration = time.Duration(1) * event.DurationScale
	eventMap[6] =event

	event = new(Event)
	event.Id = 7
	event.Name = "Event-Drive"
	event.Description = "Drive to destination"
	event.DurationScale = time.Minute
	event.Duration = time.Duration(819) * event.DurationScale
	eventMap[7] = event

	event = new(Event)
	event.Id = 8
	event.Name = "Event-Stop"
	event.Description = "Stop driving"
	event.DurationScale = time.Minute
	event.Duration = time.Duration(1) * event.DurationScale
	eventMap[8] = event

	event = new(Event)
	event.Id = 9
	event.Name = "Event-Unload"
	event.Description = "Unload the luggage"
	event.DurationScale = time.Minute
	event.Duration = time.Duration(42) * event.DurationScale
	eventMap[9] = event
}

func buildExample() {
	buildEvents()
	//computeOutgoing()
}

func main() {
	buildExample()
	dumpMap("Original list:\n", eventMap)
}