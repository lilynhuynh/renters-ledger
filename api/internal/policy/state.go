package policy

import "errors"

// Event is something that happened to a policy and may change its Status.
type Event string

const (
	EventSubmit             Event = "submit"            // customer accepts the quote
	EventPaymentSucceeded   Event = "payment_succeeded" // processor confirmed a payment
	EventPaymentFailed      Event = "payment_failed"    // declined card, returned ACH, ...
	EventGracePeriodExpired Event = "grace_period_expired"
	EventCancel             Event = "cancel"    // staff or customer cancels
	EventReinstate          Event = "reinstate" // staff reinstates a lapsed policy
)

// ErrIllegalTransition is returned when an event is not allowed in the current status.
// Callers check it with errors.Is. The HTTP layer maps it to 409 Conflict.
var ErrIllegalTransition = errors.New("illegal policy transition")

// Transition is the single source of truth for the lifecycle. It is a pure function:
// no I/O, no clock, no database, so it is trivial to test exhaustively.
//
// Legal transitions (everything else is illegal):
//
//	quoted     --submit-->               pending
//	quoted     --cancel-->               cancelled
//	pending    --payment_succeeded-->    active
//	pending    --cancel-->               cancelled
//	active     --payment_failed-->       past_due
//	active     --cancel-->               cancelled
//	past_due   --payment_succeeded-->    active
//	past_due   --grace_period_expired--> lapsed
//	past_due   --cancel-->               cancelled
//	lapsed     --reinstate-->            reinstated
//	reinstated --payment_succeeded-->    active
//	reinstated --payment_failed-->       past_due
//	reinstated --cancel-->               cancelled
//
// cancelled is terminal. lapsed accepts only reinstate.
//
// TODO(Day4): implement. On an illegal move, return (current, error) where the error
// wraps ErrIllegalTransition and names both values, for example
// fmt.Errorf("%w: %s on %s", ErrIllegalTransition, event, current).
// Hint: a map[Status]map[Event]Status table is compact, and a switch also works.
// Either way, state_test.go is the spec.
func Transition(current Status, event Event) (Status, error) {
	return "", nil
}
