// Trakt.tv importer.

package main

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"gorm.io/gorm"
)

type TraktUser struct {
	Username string `json:"username"`
	Private  bool   `json:"private"`
	IDs      struct {
		Slug string `json:"slug"`
	} `json:"ids"`
}

type TraktLists []struct {
	Name string `json:"name"`
	IDs  struct {
		Trakt string `json:"trakt"`
	} `json:"ids"`
}

type TraktImportResponse struct {
	JobId string `json:"jobId"`
}

func startTraktImport(db *gorm.DB, jobId string, userId uint, traktUsername string) {
	// Get trakt user. We want to get their profile `slug` for use in
	// next requests and we can check their profile isn't private while here.
	var traktUser TraktUser
	err := traktAPIRequest("users/"+traktUsername, &traktUser)
	if err != nil {
		slog.Error("startTraktImport: Failed to get users profile", "error", err)
		addJobError(jobId, userId, "failed to request trakt profile from api")
		updateJobStatus(jobId, userId, JOB_CANCELLED)
		return
	}
	if traktUser.Private {
		slog.Error("startTraktImport: Users profile is private. Cannot continue with import.")
		addJobError(jobId, userId, "trakt profile is private")
		updateJobStatus(jobId, userId, JOB_CANCELLED)
		return
	}
	userSlug := traktUser.IDs.Slug
	toImport := map[string]ImportRequest{}
	// Get all lists for this user
	var lists TraktLists
	err = traktAPIRequest("users/"+userSlug+"/lists", &lists)
	if err != nil {
		slog.Error("startTraktImport: Failed to get users lists", "error", err)
		addJobError(jobId, userId, "failed to get your lists")
	} else {
		for _, v := range lists {

		}
	}

	// watchlist = all PLANNED stuff?
	// History = all WATCHED stuff?
}

func traktAPIRequest(ep string, resp interface{}) error {
	slog.Debug("traktAPIRequest", "endpoint", ep)
	base, err := url.Parse("https://api.themoviedb.org/3")
	if err != nil {
		return errors.New("failed to parse api uri")
	}
	base.Path += ep

	req, err := http.NewRequest("GET", base.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Add("trakt-api-key", "c481cb044dcd58d83f3fde113741d1e28d19c1bef1bcbfcb9acedee222f3a673")
	req.Header.Add("trakt-api-version", "2")
	req.Header.Add("Content-type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return err
	}
	if !(res.StatusCode >= 200 && res.StatusCode <= 299) {
		slog.Error("traktAPIRequest: non 2xx status code:", "status_code", res.StatusCode)
		return errors.New(string(body))
	}
	err = json.Unmarshal([]byte(body), &resp)
	if err != nil {
		return err
	}
	return nil
}

func traktImportWatched(
	db *gorm.DB,
	userId uint,
	traktUsername string,
) (TraktImportResponse, error) {
	jobId, err := addJob("trakt_import", userId)
	if err != nil {
		slog.Error("traktSyncWatched: Failed to create a job", "error", err)
		return TraktImportResponse{}, errors.New("failed to create job")
	}

	updateJobStatus(jobId, userId, JOB_RUNNING)

	go startTraktImport(
		db,
		jobId,
		userId,
		traktUsername,
	)

	return TraktImportResponse{JobId: jobId}, nil
}
