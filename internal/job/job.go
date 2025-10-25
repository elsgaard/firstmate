package job

import "sync"

func Start() {
	j.start()
}

func Stop() {
	j.stop()
}

func Reset() {
	j.reset()
}

func GetStatus() bool {
	return j.status()
}

func SetStatusText(statusText string) {
	j.statusText(statusText)
}

func GetStatusText() []string {
	return j.getStatusText()
}

type job struct {
	sync.RWMutex
	running bool
	text    []string
}

func (j *job) start() {
	j.Lock()
	j.running = true
	j.Unlock()
}

func (j *job) stop() {
	j.Lock()
	j.running = false
	j.Unlock()
}

func (j *job) status() bool {
	return j.running
}

func (j *job) statusText(statusText string) {
	j.Lock()
	j.text = append(j.text, statusText)
	j.Unlock()
}

func (j *job) getStatusText() []string {
	return j.text
}

func (j *job) reset() {
	j.Lock()
	j.running = false
	j.text = []string{}
	j.Unlock()
}

var j = &job{}
