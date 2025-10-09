package govuk

import (
	"context"
	"encoding/json"
	"errors"

	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/telemetry"
)

type Holiday struct {
	Date string `json:"date"`
}

type holidayApiResponse struct {
	Container struct {
		Events []Holiday `json:"events"`
	} `json:"england-and-wales"`
}

func (c *Client) getHolidays(ctx context.Context) ([]Holiday, error) {
	logger := telemetry.LoggerFromContext(ctx)
	req, err := c.newHolidayRequest(ctx)

	if err != nil {
		logger.Error("unable to build bank holidays request", "error", err)
		return []Holiday{}, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		logger.Error("unable to send bank holidays request", "error", err)
		return []Holiday{}, err
	}

	defer unchecked(resp.Body.Close)

	if resp.StatusCode != http.StatusOK {
		logger.Error("bank holidays request returned unexpected status code", "status", resp.Status)
		return []Holiday{}, errors.New("bank holidays api error: status " + resp.Status)
	}

	var response holidayApiResponse
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return []Holiday{}, err
	}

	return response.Container.Events, nil
}

func (c *Client) isWorkingDay(d time.Time) bool {
	return !c.caches.isHoliday(d) && !c.isWeekend(d)
}

func (c *Client) isWeekend(d time.Time) bool {
	weekday := d.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

func (c *Client) AddWorkingDays(ctx context.Context, d time.Time, n int) (time.Time, error) {
	if c.caches.shouldRefreshHolidays() {
		logger := telemetry.LoggerFromContext(ctx)
		logger.Info("refreshing holidays cache via API")
		holidays, err := c.getHolidays(ctx)
		if err != nil {
			logger.Error("error in refreshing holidays cache via API", "error", err)
			return time.Time{}, err
		}
		c.caches.updateHolidays(holidays)
	}
	for {
		if n == 0 {
			return d, nil
		}
		for {
			d = d.AddDate(0, 0, 1)
			if c.isWorkingDay(d) {
				break
			}
		}
		n--
	}
}

// NextWorkingDayOnOrAfterX will return the next available working day on or after dayOfMonth from date. e.g. if dayOfMonth
// is 24 and is a working day, the date returned will be 24th of date's current month if that date has not passed, otherwise
// the 24th of the next month.
func (c *Client) NextWorkingDayOnOrAfterX(ctx context.Context, date time.Time, dayOfMonth int) (time.Time, error) {
	if c.caches.shouldRefreshHolidays() {
		logger := telemetry.LoggerFromContext(ctx)
		logger.Info("refreshing holidays cache via API")
		holidays, err := c.getHolidays(ctx)
		if err != nil {
			logger.Error("error in refreshing holidays cache via API", "error", err)
			return time.Time{}, err
		}
		c.caches.updateHolidays(holidays)
	}

	if date.Day() > dayOfMonth {
		date = date.AddDate(0, 1, 0)
	}

	date = time.Date(date.Year(), date.Month(), dayOfMonth, 0, 0, 0, 0, time.UTC)

	for {
		if b := c.isWorkingDay(date); b {
			return date, nil
		}
		date = date.AddDate(0, 0, 1)
	}
}
