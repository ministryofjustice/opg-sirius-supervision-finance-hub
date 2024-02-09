package server

import (
	"fmt"
	"github.com/opg-sirius-finance-hub/internal/model"
	"github.com/opg-sirius-finance-hub/internal/sirius"
	"golang.org/x/sync/errgroup"
	"net/http"
	"regexp"
	"strconv"
)

type FinanceVars struct {
	Path            string
	XSRFToken       string
	MyDetails       model.Assignee
	Person          model.Person
	Tabs            []Tab
	SuccessMessage  string
	Errors          sirius.ValidationErrors
	EnvironmentVars EnvironmentVars
}

type Tab struct {
	Title    string
	BasePath string
}

type FinanceVarsClient interface {
	GetCurrentUserDetails(sirius.Context) (model.Assignee, error)
	GetPersonDetails(sirius.Context, int) (model.Person, error)
}

func NewFinanceVars(client FinanceVarsClient, r *http.Request, envVars EnvironmentVars) (*FinanceVars, error) {
	ctx := getContext(r)
	group, groupCtx := errgroup.WithContext(ctx.Context)

	vars := FinanceVars{
		Path:            r.URL.Path,
		XSRFToken:       ctx.XSRFToken,
		EnvironmentVars: envVars,
		Tabs: []Tab{
			{
				Title:    "Invoices",
				BasePath: "invoices",
			},
		},
	}

	group.Go(func() error {
		myDetails, err := client.GetCurrentUserDetails(ctx.With(groupCtx))
		if err != nil {
			return err
		}
		vars.MyDetails = myDetails
		return nil
	})
	group.Go(func() error {
		re := regexp.MustCompile("[0-9]+")
		fmt.Println(r.URL.Path)
		personId, _ := strconv.Atoi(re.FindAllString(r.URL.Path, -1)[0])
		person, err := client.GetPersonDetails(ctx.With(groupCtx), personId)
		if err != nil {
			return err
		}
		vars.Person = person
		return nil
	})

	if err := group.Wait(); err != nil {
		return nil, err
	}

	return &vars, nil
}

func (t Tab) GetURL(personId int) string {
	return "/" + strconv.Itoa(personId)
}
