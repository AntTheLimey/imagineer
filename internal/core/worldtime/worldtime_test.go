/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package worldtime

import (
	"testing"
	"time"
)

// daysSinceEpochStdlib counts days from 0001-01-01 to the given civil date
// using Go's standard library, which also uses the proleptic Gregorian
// calendar. Unix seconds are used rather than time.Sub because a Duration
// overflows beyond roughly 292 years.
func daysSinceEpochStdlib(y, m, d int) int64 {
	base := time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	at := time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC).Unix()
	return (at - base) / 86400
}

func TestGregorianToHours(t *testing.T) {
	// RD(0001-01-01) = 1 → hour 0.
	if got := GregorianToHours(1, 1, 1, 0); got != 0 {
		t.Fatalf("epoch = %d", got)
	}
	// RD(1814-08-10) = 662406 (365*1813 + 453 − 18 + 4 + 222).
	// Dawn (06:00) = 662405*24 + 6 = 15_897_726.
	if got := GregorianToHours(1814, 8, 10, 6); got != 15_897_726 {
		t.Fatalf("duel dawn = %d", got)
	}
	// Leap-century check: 2000-03-01 follows Feb 29.
	if GregorianToHours(2000, 3, 1, 0)-GregorianToHours(2000, 2, 29, 0) != 24 {
		t.Fatal("2000 leap day wrong")
	}
	// 1900 is NOT a leap year.
	if GregorianToHours(1900, 3, 1, 0)-GregorianToHours(1900, 2, 28, 0) != 24 {
		t.Fatal("1900 must not be leap")
	}
}

func TestRoundTrip(t *testing.T) {
	for _, h := range []int64{0, 15_897_726, -8760, 123_456_789} {
		y, m, d, hr := HoursToGregorian(h)
		if back := GregorianToHours(y, m, d, hr); back != h {
			t.Fatalf("round trip %d → %d-%d-%d %d → %d", h, y, m, d, hr, back)
		}
	}
}

func TestDayRange(t *testing.T) {
	r := DayRange(1814, 8, 10)
	if r.Lo != 15_897_720 || r.Hi != 15_897_744 {
		t.Fatalf("day range = %+v", r)
	}
}

// ---------------------------------------------------------------------------
// Additional adversarial coverage beyond the brief.
// ---------------------------------------------------------------------------

// TestHoursToGregorianValues pins the decoded calendar date, not merely the
// round-trip identity. A conversion pair can be self-consistent and still be
// wrong; these assertions fix the absolute answer.
func TestHoursToGregorianValues(t *testing.T) {
	cases := []struct {
		h                  int64
		y, m, d, hourOfDay int
	}{
		{0, 1, 1, 1, 0},
		{23, 1, 1, 1, 23},
		{24, 1, 1, 2, 0},
		{15_897_726, 1814, 8, 10, 6},
		{-1, 0, 12, 31, 23},     // last hour before the epoch
		{-24, 0, 12, 31, 0},     // start of 0000-12-31
		{-8760, 0, 1, 2, 0},     // brief's negative case: 366-day year 0
		{-8784, 0, 1, 1, 0},     // 0000-01-01, i.e. RD −365
		{-8785, -1, 12, 31, 23}, // last hour of year −1
		{-17544, -1, 1, 1, 0},   // start of year −1, i.e. (rataDie(−1,1,1)−1)*24
	}
	for _, c := range cases {
		y, m, d, hr := HoursToGregorian(c.h)
		if y != c.y || m != c.m || d != c.d || hr != c.hourOfDay {
			t.Errorf("HoursToGregorian(%d) = %d-%02d-%02d %02d:00, want %d-%02d-%02d %02d:00",
				c.h, y, m, d, hr, c.y, c.m, c.d, c.hourOfDay)
		}
		if back := GregorianToHours(y, m, d, hr); back != c.h {
			t.Errorf("GregorianToHours(%d-%02d-%02d %02d) = %d, want %d",
				y, m, d, hr, back, c.h)
		}
	}
}

// TestPreEpochYearLengths is the regression test for the floor-division fix in
// rataDie. With truncating division the base term for year ≤ 0 is wrong, which
// makes year 0 come out 365 days long and aliases 0000-12-31 onto 0001-01-01.
func TestPreEpochYearLengths(t *testing.T) {
	// Year 0 is a leap year (0 % 400 == 0) and must be 366 days.
	if got := GregorianToHours(1, 1, 1, 0) - GregorianToHours(0, 1, 1, 0); got != 366*24 {
		t.Errorf("length of year 0 = %d hours, want %d", got, 366*24)
	}
	// The day before the epoch is 0000-12-31, not 0001-01-01.
	if got := GregorianToHours(1, 1, 1, 0) - GregorianToHours(0, 12, 31, 0); got != 24 {
		t.Errorf("0000-12-31 → 0001-01-01 = %d hours, want 24", got)
	}
	// Year −1 (2 BCE) is not a leap year.
	if got := GregorianToHours(0, 1, 1, 0) - GregorianToHours(-1, 1, 1, 0); got != 365*24 {
		t.Errorf("length of year -1 = %d hours, want %d", got, 365*24)
	}
	// Year −4 is a leap year; year −100 is not; year −400 is.
	for _, c := range []struct {
		y    int
		days int64
	}{{-4, 366}, {-100, 365}, {-400, 366}, {-401, 365}} {
		got := GregorianToHours(c.y+1, 1, 1, 0) - GregorianToHours(c.y, 1, 1, 0)
		if got != c.days*24 {
			t.Errorf("length of year %d = %d hours, want %d", c.y, got, c.days*24)
		}
	}
}

// TestDecemberDecodes is the regression test for the cumDays bounds fix. The
// month walk sizes month m as cumDays[m] − cumDays[m-1]; without a trailing
// sentinel entry, reaching December indexes past the end of the array and
// every December date panics.
func TestDecemberDecodes(t *testing.T) {
	for _, y := range []int{-1, 0, 1, 1900, 2000, 2023, 2024} {
		for d := 1; d <= 31; d++ {
			h := GregorianToHours(y, 12, d, 12)
			gy, gm, gd, ghr := HoursToGregorian(h)
			if gy != y || gm != 12 || gd != d || ghr != 12 {
				t.Errorf("%d-12-%02d 12:00 → %d-%02d-%02d %02d:00", y, d, gy, gm, gd, ghr)
			}
		}
	}
}

// TestMonotonicAndInjective walks every day across the epoch boundary and
// asserts the hours line advances by exactly 24 per day. Any aliasing or gap in
// the Rata Die inversion shows up here.
func TestMonotonicAndInjective(t *testing.T) {
	prev := GregorianToHours(-2, 1, 1, 0)
	for h := prev + 24; h <= GregorianToHours(4, 1, 1, 0); h += 24 {
		y, m, d, hr := HoursToGregorian(h)
		if hr != 0 {
			t.Fatalf("hour %d decoded hour-of-day %d, want 0", h, hr)
		}
		back := GregorianToHours(y, m, d, hr)
		if back != h {
			t.Fatalf("hour %d → %d-%02d-%02d → %d", h, y, m, d, back)
		}
		if back-prev != 24 {
			t.Fatalf("gap at %d-%02d-%02d: %d hours since previous day", y, m, d, back-prev)
		}
		prev = back
	}
}

// TestBruteForceRoundTrip sweeps a window spanning the epoch, exercising the
// floor semantics of the hour extraction for negative inputs.
func TestBruteForceRoundTrip(t *testing.T) {
	for h := int64(-100_000); h <= 100_000; h += 137 {
		y, m, d, hr := HoursToGregorian(h)
		if hr < 0 || hr > 23 {
			t.Fatalf("HoursToGregorian(%d) hour-of-day = %d, out of range", h, hr)
		}
		if m < 1 || m > 12 {
			t.Fatalf("HoursToGregorian(%d) month = %d, out of range", h, m)
		}
		if d < 1 || d > 31 {
			t.Fatalf("HoursToGregorian(%d) day = %d, out of range", h, d)
		}
		if back := GregorianToHours(y, m, d, hr); back != h {
			t.Fatalf("round trip %d → %d-%02d-%02d %02d → %d", h, y, m, d, hr, back)
		}
	}
}

// TestWideRoundTrip covers far-future and deep-past values where the year
// search seed matters.
func TestWideRoundTrip(t *testing.T) {
	for _, h := range []int64{
		-100_000_000, -15_897_726, -1_000_000, -87_660, -25, -24, -23, -1,
		1, 23, 25, 8760, 1_000_000, 123_456_789, 100_000_000,
	} {
		y, m, d, hr := HoursToGregorian(h)
		if back := GregorianToHours(y, m, d, hr); back != h {
			t.Fatalf("round trip %d → %d-%02d-%02d %02d → %d", h, y, m, d, hr, back)
		}
	}
}

// TestMonthLengths checks that consecutive month starts differ by exactly the
// number of days in the month, for a leap year and a non-leap year.
func TestMonthLengths(t *testing.T) {
	common := [12]int64{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	leap := [12]int64{31, 29, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}

	check := func(y int, want [12]int64) {
		t.Helper()
		for m := 1; m <= 12; m++ {
			ny, nm := y, m+1
			if nm == 13 {
				ny, nm = y+1, 1
			}
			got := GregorianToHours(ny, nm, 1, 0) - GregorianToHours(y, m, 1, 0)
			if got != want[m-1]*24 {
				t.Errorf("year %d month %d = %d hours, want %d", y, m, got, want[m-1]*24)
			}
			// The last day of the month must be exactly 24 hours before the
			// first of the next month.
			last := GregorianToHours(y, m, int(want[m-1]), 0)
			if GregorianToHours(ny, nm, 1, 0)-last != 24 {
				t.Errorf("year %d month %d last day is not day %d", y, m, want[m-1])
			}
		}
	}

	check(2023, common) // non-leap
	check(2024, leap)   // leap
	check(1900, common) // century, not leap
	check(2000, leap)   // 400-year century, leap
	check(0, leap)      // pre-epoch leap year
	check(-1, common)   // pre-epoch non-leap year
}

// TestAgainstStdlib cross-checks the Rata Die arithmetic against Go's own
// civil-date implementation over a range where time.Time is valid.
func TestAgainstStdlib(t *testing.T) {
	// time.Date uses the proleptic Gregorian calendar, so the day count
	// between any two dates is an independent check on rataDie.
	base := GregorianToHours(1, 1, 1, 0)
	for _, c := range []struct{ y, m, d int }{
		{1, 1, 1}, {1582, 10, 15}, {1814, 8, 10}, {1970, 1, 1},
		{2000, 2, 29}, {2024, 12, 31}, {9999, 12, 31},
	} {
		want := daysSinceEpochStdlib(c.y, c.m, c.d) * 24
		if got := GregorianToHours(c.y, c.m, c.d, 0) - base; got != want {
			t.Errorf("%d-%02d-%02d: %d hours since epoch, want %d", c.y, c.m, c.d, got, want)
		}
	}
}

func TestDayRangeHalfOpen(t *testing.T) {
	r := DayRange(2024, 2, 29)
	if r.Hi-r.Lo != 24 {
		t.Fatalf("DayRange span = %d, want 24", r.Hi-r.Lo)
	}
	if r.Lo != GregorianToHours(2024, 2, 29, 0) {
		t.Fatalf("DayRange.Lo = %d", r.Lo)
	}
	// Hi is exclusive: it is the first hour of the next day.
	if r.Hi != GregorianToHours(2024, 3, 1, 0) {
		t.Fatalf("DayRange.Hi = %d, want start of next day", r.Hi)
	}
	// Pre-epoch day ranges behave identically.
	pr := DayRange(0, 1, 1)
	if pr.Hi-pr.Lo != 24 || pr.Lo != -8784 {
		t.Fatalf("pre-epoch DayRange = %+v", pr)
	}
}

func TestRangeContains(t *testing.T) {
	r := DayRange(1814, 8, 10)
	for _, c := range []struct {
		h    int64
		want bool
	}{
		{r.Lo - 1, false},
		{r.Lo, true},      // Lo is inclusive
		{r.Lo + 23, true}, // last hour of the day
		{r.Hi, false},     // Hi is exclusive
		{r.Hi + 1, false},
	} {
		if got := r.Contains(c.h); got != c.want {
			t.Errorf("Range%+v.Contains(%d) = %v, want %v", r, c.h, got, c.want)
		}
	}
}

func TestCalendarTypes(t *testing.T) {
	c := Calendar{
		Scheme: SchemeGregorianRataDie,
		Months: []Month{{Name: "January", Days: 31}},
		Weekdays: []Weekday{
			{Name: "Monday", Hours: 24},
		},
	}
	if c.Scheme != "gregorian-rata-die" {
		t.Fatalf("scheme = %q", c.Scheme)
	}
	if c.Months[0].Days != 31 || c.Weekdays[0].Hours != 24 {
		t.Fatalf("calendar fields = %+v", c)
	}
}
