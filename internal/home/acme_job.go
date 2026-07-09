package home

import (
	"sync"
	"time"
)

// acmeJobLine is a single progress line recorded by an [acmeJob], suitable
// for JSON encoding as a Server-Sent Events "line" event.
type acmeJobLine struct {
	Time    time.Time `json:"time"`
	Level   string    `json:"level"`
	Message string    `json:"message"`
}

// acmeJobResult is the outcome of a finished [acmeJob], suitable for JSON
// encoding as a Server-Sent Events "done" event.
type acmeJobResult struct {
	Status           *tlsConfigStatus `json:"status,omitempty"`
	CertificateChain string           `json:"certificate_chain,omitempty"`
	PrivateKey       string           `json:"private_key,omitempty"`
	Error            string           `json:"error,omitempty"`
	Success          bool             `json:"success"`
}

// acmeJob tracks the progress of a single, in-flight ACME certificate
// issuance, and lets any number of Server-Sent Events subscribers observe it
// in real time.  A new subscriber immediately receives all lines recorded so
// far, then further lines as they're recorded.  It's safe for concurrent
// use.
type acmeJob struct {
	mu     sync.Mutex
	lines  []acmeJobLine
	notify chan struct{}
	done   bool
	result *acmeJobResult
}

// newAcmeJob returns a new, empty *acmeJob.
func newAcmeJob() (j *acmeJob) {
	return &acmeJob{notify: make(chan struct{})}
}

// log records a progress line.  It's safe for concurrent use, including
// concurrent use with other methods of j.
func (j *acmeJob) log(level, msg string) {
	j.mu.Lock()
	defer j.mu.Unlock()

	if j.done {
		return
	}

	j.lines = append(j.lines, acmeJobLine{Time: time.Now(), Level: level, Message: msg})

	close(j.notify)
	j.notify = make(chan struct{})
}

// finish marks the job as done with the given result.  Subsequent calls to
// log or finish are no-ops.  It's safe for concurrent use, including
// concurrent use with other methods of j.
func (j *acmeJob) finish(result *acmeJobResult) {
	j.mu.Lock()
	defer j.mu.Unlock()

	if j.done {
		return
	}

	j.done = true
	j.result = result

	close(j.notify)
}

// snapshot returns a copy of the lines recorded so far, a channel that's
// closed the next time new state is available (either a new line, or the job
// finishing), whether the job is done, and its result if so.  It's safe for
// concurrent use.
func (j *acmeJob) snapshot() (lines []acmeJobLine, notify chan struct{}, done bool, result *acmeJobResult) {
	j.mu.Lock()
	defer j.mu.Unlock()

	lines = append([]acmeJobLine(nil), j.lines...)

	return lines, j.notify, j.done, j.result
}

// isDone returns true once the job has finished.  It's safe for concurrent
// use.
func (j *acmeJob) isDone() (done bool) {
	j.mu.Lock()
	defer j.mu.Unlock()

	return j.done
}
