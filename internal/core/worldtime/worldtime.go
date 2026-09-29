/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

// Package worldtime represents in-world time as a single numeric line
// measured in hours since a calendar epoch.
//
// For worlds that use Earth history the epoch is the Rata Die day number,
// where day 1 is 1 January 1 CE in the proleptic Gregorian calendar. Hour 0
// is therefore 0001-01-01 00:00 and the conversion is:
//
//	hours = (RD − 1) * 24 + hourOfDay
//
// Negative hours are legitimate and denote pre-epoch times; every function
// here uses floor semantics so that the line is continuous and strictly
// monotonic across the epoch.
//
// Fuzzy or imprecise world times are expressed as a wide half-open interval
// (see Range), not as a single instant. Invented calendars describe
// themselves with Calendar, whose days need not be 24 hours long.
package worldtime

// Month is a named month of an invented or historical calendar.
type Month struct {
	Name string
	Days int
}

// Weekday is a named day of an invented or historical calendar. Hours records
// the length of the day, which need not be 24 for invented calendars.
type Weekday struct {
	Name  string
	Hours int
}

// Calendar describes how a world's dates are named and subdivided. The Scheme
// selects the arithmetic: SchemeGregorianRataDie means the conversions in this
// package apply directly, while SchemeCustom means Months and Weekdays define
// the calendar entirely.
type Calendar struct {
	Scheme   string
	Months   []Month
	Weekdays []Weekday
}

// Recognised Calendar.Scheme values.
const (
	// SchemeGregorianRataDie is the proleptic Gregorian calendar counted from
	// the Rata Die epoch, as implemented by GregorianToHours.
	SchemeGregorianRataDie = "gregorian-rata-die"
	// SchemeCustom is an invented calendar defined by its Months and Weekdays.
	SchemeCustom = "custom"
)

// HoursPerDay is the length of a Gregorian day on the hours line.
const HoursPerDay = 24

// Range is a half-open interval [Lo, Hi) on the hours line. It is the
// representation of a fuzzy world time: the wider the interval, the less
// precisely the event is placed.
type Range struct{ Lo, Hi int64 }

// Contains reports whether h falls within the half-open interval.
func (r Range) Contains(h int64) bool { return h >= r.Lo && h < r.Hi }

// cumDays[m] is the number of days in a non-leap year before month m+1
// begins. The trailing 365 is a sentinel so that month 12 can be sized as
// cumDays[12] − cumDays[11] without a special case.
var cumDays = [...]int{0, 31, 59, 90, 120, 151, 181, 212, 243, 273, 304, 334, 365}

// isLeap reports whether y is a leap year in the proleptic Gregorian calendar.
// Go's % keeps the sign of the dividend, but since the tests here are all
// against zero this behaves correctly for negative (pre-1 CE) years too.
func isLeap(y int) bool {
	return y%4 == 0 && (y%100 != 0 || y%400 == 0)
}

// daysInMonth returns the length of month m (1..12) in year y.
func daysInMonth(y, m int) int {
	n := cumDays[m] - cumDays[m-1]
	if m == 2 && isLeap(y) {
		n++
	}
	return n
}

// floorDiv divides a by b rounding towards negative infinity. Go's integer
// division truncates towards zero, which yields the wrong leap-year counts for
// pre-epoch years, so the Rata Die arithmetic must not use / directly.
func floorDiv(a, b int64) int64 {
	q := a / b
	if a%b != 0 && (a < 0) != (b < 0) {
		q--
	}
	return q
}

// rataDie returns the Rata Die day number of the proleptic Gregorian date
// y-m-d, where 0001-01-01 is day 1. Years may be zero or negative; month must
// be in 1..12.
func rataDie(y, m, d int) int64 {
	yy := int64(y - 1)
	rd := 365*yy + floorDiv(yy, 4) - floorDiv(yy, 100) + floorDiv(yy, 400) +
		int64(cumDays[m-1]) + int64(d)
	if m > 2 && isLeap(y) {
		rd++
	}
	return rd
}

// GregorianToHours converts a proleptic Gregorian date and hour-of-day to the
// number of hours since the Rata Die epoch (0001-01-01 00:00 = 0). Results are
// negative for pre-epoch dates. Month must be in 1..12.
func GregorianToHours(y, m, d, hour int) int64 {
	return (rataDie(y, m, d)-1)*HoursPerDay + int64(hour)
}

// HoursToGregorian is the exact inverse of GregorianToHours: it converts a
// point on the hours line back to a proleptic Gregorian date and hour-of-day.
// The returned hour is always in 0..23, including for negative h.
func HoursToGregorian(h int64) (y, m, d, hour int) {
	// Floor-divide into whole days plus a non-negative hour-of-day.
	days := h / HoursPerDay
	hour = int(h - days*HoursPerDay)
	if hour < 0 {
		hour += HoursPerDay
		days--
	}
	rd := days + 1

	// Seed the year from the shortest possible year length, then correct in
	// whichever direction is needed. Floor division keeps the seed close for
	// negative Rata Die values, where truncation would overshoot towards zero.
	y = int(floorDiv(rd, 366)) + 1
	for rataDie(y+1, 1, 1) <= rd {
		y++
	}
	for rataDie(y, 1, 1) > rd {
		y--
	}

	// Walk the months of that year.
	doy := int(rd - rataDie(y, 1, 1) + 1)
	m = 1
	for {
		dim := daysInMonth(y, m)
		if m == 12 || doy <= dim {
			break
		}
		doy -= dim
		m++
	}
	return y, m, doy, hour
}

// DayRange returns the whole-day fuzzy interval for the given Gregorian date:
// the half-open range covering all 24 hours of that day.
func DayRange(y, m, d int) Range {
	lo := GregorianToHours(y, m, d, 0)
	return Range{Lo: lo, Hi: lo + HoursPerDay}
}
