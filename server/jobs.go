// We will use jobs only for storing/retrieving active job statuses for the client.
// Running the job will be done wherever needed, but isn't handled here.
// When starting a job elsewhere, we should first add a job here as active to get an `id`,
// this id should be used to update the active job so the client can request job status updates.

package main

import (
	"errors"
	"log/slog"
)

type JobStatus string

var (
	JOB_CREATED   JobStatus = "CREATED"
	JOB_RUNNING   JobStatus = "RUNNING"
	JOB_COMPLETED JobStatus = "COMPLETED"
	JOB_FAILED    JobStatus = "FAILED"
)

type Job struct {
	// We can give the job a name simply for showing on the client.
	Name string `json:"name"`
	// The current status of the job.
	Status JobStatus `json:"status"`
	// The current task we are performing inside the job.
	// Just so we can portray progress on the client by displaying the current task.
	CurrentTask string `json:"currentTask,omitempty"`
	// Errors that occurred in the task
	Errors []string `json:"errors"`
	// Stored for access control.
	UserId uint `json:"-"`
}

var activeJobs = make(map[string]*Job)

// Add a job to our activeJobs map.
// Returns id of job on success, or error if failed to add.
func addJob(name string, userId uint) (string, error) {
	idk, err := generateString(8)
	if err != nil {
		return "", err
	}
	_, ok := activeJobs[idk]
	if ok {
		// Lets just hope this doesn't happen, may the odds be with us.
		return "", errors.New("job already exists with id generated, please try again")
	}
	activeJobs[idk] = &Job{
		Name:   name,
		Status: JOB_CREATED,
		UserId: userId,
	}
	return idk, nil
}

// Get a job.
// Returns job if found, otherwise errors if job does not exist.
func getJob(id string, userId uint) (*Job, error) {
	j, ok := activeJobs[id]
	if ok {
		// Ensure user requesting a job, owns the job.
		if j.UserId != userId {
			slog.Warn("getJob: A user tried to access a job they do not own.", "user_id", userId, "job_id", id)
			return &Job{}, errors.New("job does not exist")
		}
		return j, nil
	}
	return &Job{}, errors.New("job does not exist")
}

// Update a jobs status.
func updateJobStatus(id string, userId uint, status JobStatus) error {
	j, err := getJob(id, userId)
	if err != nil {
		slog.Error("updateJobStatus: Failed!", "status", status, "error", err)
		return err
	}
	j.Status = status
	return nil
}

// Update a jobs current task.
func updateJobCurrentTask(id string, userId uint, ct string) error {
	j, err := getJob(id, userId)
	if err != nil {
		slog.Error("updateJobCurrentTask: Failed!", "ct", ct, "error", err)
		return err
	}
	j.CurrentTask = ct
	return nil
}

// Add an error to a job.
func addJobError(id string, userId uint, e string) error {
	j, err := getJob(id, userId)
	if err != nil {
		slog.Error("updateJobCurrentTask: Failed!", "e", e, "error", err)
		return err
	}
	j.Errors = append(j.Errors, e)
	return nil
}
