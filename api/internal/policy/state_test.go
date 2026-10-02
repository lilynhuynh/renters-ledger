package policy

import (
	"errors"
	"testing"
)

func TestTransition(t *testing.T) {
	t.Skip("Day4: remove this line when you implement Transition")

	tests := []struct {
		from    Status
		event   Event
		want    Status
		wantErr bool
	}{
		// --- legal ---
		{StatusQuoted, EventSubmit, StatusPending, false},
		{StatusQuoted, EventCancel, StatusCancelled, false},
		{StatusPending, EventPaymentSucceeded, StatusActive, false},
		{StatusPending, EventCancel, StatusCancelled, false},
		{StatusActive, EventPaymentFailed, StatusPastDue, false},
		{StatusActive, EventCancel, StatusCancelled, false},
		{StatusPastDue, EventPaymentSucceeded, StatusActive, false},
		{StatusPastDue, EventGracePeriodExpired, StatusLapsed, false},
		{StatusPastDue, EventCancel, StatusCancelled, false},
		{StatusLapsed, EventReinstate, StatusReinstated, false},
		{StatusReinstated, EventPaymentSucceeded, StatusActive, false},
		{StatusReinstated, EventPaymentFailed, StatusPastDue, false},
		{StatusReinstated, EventCancel, StatusCancelled, false},

		// --- illegal: status must stay the same and the error must wrap ErrIllegalTransition ---
		{StatusQuoted, EventPaymentSucceeded, StatusQuoted, true},
		{StatusPending, EventSubmit, StatusPending, true},
		{StatusActive, EventReinstate, StatusActive, true},
		{StatusActive, EventGracePeriodExpired, StatusActive, true},
		{StatusLapsed, EventPaymentSucceeded, StatusLapsed, true},
		{StatusLapsed, EventCancel, StatusLapsed, true},
		{StatusCancelled, EventReinstate, StatusCancelled, true},
		{StatusCancelled, EventPaymentSucceeded, StatusCancelled, true},
		{StatusCancelled, EventCancel, StatusCancelled, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.from)+"/"+string(tt.event), func(t *testing.T) {
			got, err := Transition(tt.from, tt.event)
			if tt.wantErr {
				if !errors.Is(err, ErrIllegalTransition) {
					t.Fatalf("err = %v, want ErrIllegalTransition", err)
				}
			} else if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if got != tt.want {
				t.Errorf("Transition(%s, %s) = %s, want %s", tt.from, tt.event, got, tt.want)
			}
		})
	}
}
