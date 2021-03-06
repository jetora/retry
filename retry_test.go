package retry

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

var delta = 10 * time.Millisecond

func TestRetry(t *testing.T) {
	type Assert struct {
		Attempts uint
		Error    func(error) bool
	}

	tests := map[string]struct {
		breaker    BreakCloser
		strategies []func(attempt uint, err error) bool
		error      error
		assert     Assert
	}{
		"zero iterations": {
			newClosedBreaker(),
			[]func(attempt uint, err error) bool{delay(delta), limit(10000)},
			errors.New("zero iterations"),
			Assert{0, func(err error) bool { return IsInterrupted(err) && err.Error() == string(Interrupted) }},
		},
		"one iteration": {
			newBreaker(),
			nil,
			nil,
			Assert{1, func(err error) bool { return err == nil }},
		},
		"two iterations": {
			newBreaker(),
			[]func(attempt uint, err error) bool{limit(2)},
			errors.New("two iterations"),
			Assert{2, func(err error) bool { return err != nil && err.Error() == "two iterations" }},
		},
		"three iterations": {
			newBreaker(),
			[]func(attempt uint, err error) bool{limit(3)},
			errors.New("three iterations"),
			Assert{3, func(err error) bool { return err != nil && err.Error() == "three iterations" }},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var total uint
			action := func(attempt uint) error {
				total = attempt + 1
				return test.error
			}
			err := Retry(test.breaker, action, test.strategies...)
			if !test.assert.Error(err) {
				t.Error("fail error assertion")
			}
			if _, is := IsRecovered(err); is {
				t.Error("recovered panic is not expected")
			}
			if test.assert.Attempts != total {
				t.Errorf("expected %d attempts, obtained %d", test.assert.Attempts, total)
			}
		})
	}
	t.Run("unexpected panic", func(t *testing.T) {
		err := Retry(newBreaker(), func(uint) error { panic("Catch Me If You Can") })
		cause, is := IsRecovered(err)
		if !is {
			t.Fatal("recovered panic is expected")
		}
		if !reflect.DeepEqual(cause, "Catch Me If You Can") {
			t.Fatal("Catch Me If You Can is expected")
		}
	})
}

func TestTry(t *testing.T) {
	type Assert struct {
		Attempts uint
		Error    func(error) bool
	}

	tests := map[string]struct {
		breaker    Breaker
		strategies []func(attempt uint, err error) bool
		error      error
		assert     Assert
	}{
		"zero iterations": {
			newClosedBreaker(),
			[]func(attempt uint, err error) bool{delay(delta), limit(10000)},
			errors.New("zero iterations"),
			Assert{0, func(err error) bool { return IsInterrupted(err) && err.Error() == string(Interrupted) }},
		},
		"one iteration": {
			newBreaker(),
			nil,
			nil,
			Assert{1, func(err error) bool { return err == nil }},
		},
		"two iterations": {
			newBreaker(),
			[]func(attempt uint, err error) bool{limit(2)},
			errors.New("two iterations"),
			Assert{2, func(err error) bool { return err != nil && err.Error() == "two iterations" }},
		},
		"three iterations": {
			newPanicBreaker(),
			[]func(attempt uint, err error) bool{limit(3)},
			errors.New("three iterations"),
			Assert{3, func(err error) bool { return err != nil && err.Error() == "three iterations" }},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var total uint
			action := func(attempt uint) error {
				total = attempt + 1
				return test.error
			}
			err := Try(test.breaker, action, test.strategies...)
			if !test.assert.Error(err) {
				t.Error("fail error assertion")
			}
			if _, is := IsRecovered(err); is {
				t.Error("recovered panic is not expected")
			}
			if test.assert.Attempts != total {
				t.Errorf("expected %d attempts, obtained %d", test.assert.Attempts, total)
			}
		})
	}
	t.Run("unexpected panic", func(t *testing.T) {
		err := Try(newBreaker(), func(uint) error { panic("Catch Me If You Can") })
		cause, is := IsRecovered(err)
		if !is {
			t.Fatal("recovered panic is expected")
		}
		if !reflect.DeepEqual(cause, "Catch Me If You Can") {
			t.Fatal("Catch Me If You Can is expected")
		}
	})
}
