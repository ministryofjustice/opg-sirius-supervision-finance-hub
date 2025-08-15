package api

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"net/http"
	"time"
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

func (c *Client) addWorkingDays(ctx context.Context, d time.Time, n int) (time.Time, error) {
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

func (c *Client) lastWorkingDayOfMonth(ctx context.Context, d time.Time) (time.Time, error) {
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

	d = time.Date(d.Year(), d.Month()+1, 0, 0, 0, 0, 0, time.UTC) // day 0 will underflow to the end of the previous month

	for {
		if b := c.caches.isHoliday(d); !b {
			return d, nil
		}
		d = d.AddDate(0, 0, -1)
	}
}
