/*
Copyright 2015 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// An external diurnal controller for kubernetes. With this, it's possible to manage
// known replica counts that vary throughout the day.

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	kclient "github.com/GoogleCloudPlatform/kubernetes/pkg/client"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/labels"

	"github.com/golang/glog"
)

const dayPeriod = 24 * time.Hour

type timeCount struct {
	time  time.Duration
	count int
}

func (tc timeCount) String() string {
	h := tc.time / time.Hour
	m := (tc.time % time.Hour) / time.Minute
	s := (tc.time % time.Minute) / time.Second
	if m == 0 && s == 0 {
		return fmt.Sprintf("(%02d, %d)", h, tc.count)
	} else if s == 0 {
		return fmt.Sprintf("(%02d:%02d, %d)", h, m, tc.count)
	}
	return fmt.Sprintf("(%02d:%02d:%02d, %d)", h, m, s, tc.count)
}

func timeMustParse(layout, s string) time.Time {
	t, err := time.Parse(layout, s)
	if err != nil {
		panic(err)
	}
	return t
}

var epoch = timeMustParse("150405", "000000")

func parseTimeRelative(s string) (time.Duration, error) {
	layouts := []string{"150405", "15:04:05", "1504", "15:04", "15"}
	for _, layout := range layouts {
		t, err := time.Parse(layout, s)
		if err == nil {
			return t.Sub(epoch), nil
		}
	}
	return 0, fmt.Errorf("%s must be in the form hhmmss, hh:mm:ss, hhmm, hh:mm, or hh", s)
}

func parseTimeCounts(times string, counts string) ([]timeCount, error) {
	ts := strings.Split(times, ",")
	cs := strings.Split(counts, ",")
	if len(ts) != len(cs) {
		return nil, fmt.Errorf("provided %d times but %d replica counts", len(ts), len(cs))
	}
	var tc []timeCount
	prev := time.Duration(-1)
	for i := range ts {
		t, err := parseTimeRelative(ts[i])
		if err != nil {
			return nil, err
		}
		if t < 0 {
			return nil, errors.New("times must be non-negative")
		}
		if t <= prev {
			return nil, errors.New("times must be in increasing order")
		}
		prev = t
		c, err := strconv.ParseInt(cs[i], 10, 64)
		if c < 0 {
			return nil, errors.New("counts must be non-negative")
		}
		if err != nil {
			return nil, err
		}
		tc = append(tc, timeCount{t, int(c)})
	}
	return tc, nil
}

type Scaler struct {
	timeCounts []timeCount
	selector   labels.Selector
	start      time.Time
	pos        int
	done       chan struct{}
}

func NewScaler(tc []timeCount, selector labels.Selector) (*Scaler, error) {
	smoothed, err := smoothChanges(tc, *maxScaleRate)
	if err != nil {
		return nil, err
	}
	return &Scaler{timeCounts: smoothed, selector: selector}, nil
}

// smoothChanges will add intermediate changes so that the rate of change will not
// exceed maxRate (number of replicas added/removed per minute). It does not linearly
// interpolate the changes, only insert additional steps before the change so that
// maxRate is not exceeded.
func smoothChanges(tc []timeCount, maxRate float64) ([]timeCount, error) {
	if len(tc) == 0 {
		return tc, nil
	}
	// Need to smooth the end as well
	tc = append(tc, tc[0])
	smoothed := []timeCount{tc[0]}
	for i := 1; i < len(tc); i++ {
		tdiff := float64(tc[i].time-tc[i-1].time) / float64(time.Minute)
		if i == len(tc)-1 {
			tdiff = float64(dayPeriod+tc[i-1].time-tc[i].time) / float64(time.Minute)
		}
		cdiff := float64(tc[i].count - tc[i-1].count)
		sign := 1.0
		if cdiff < 0 {
			cdiff = -cdiff
			sign = -1
		}
		rate := cdiff / tdiff
		if rate > maxRate {
			return nil, errors.New("cannot meet maximum scaling rate with given times and replica counts")
		}
		var intermediateCounts []int
		count := float64(tc[i].count) - sign*maxRate
		for sign*count > sign*float64(tc[i-1].count) {
			intermediateCounts = append(intermediateCounts, int(count))
			count -= sign * maxRate
		}
		for j := len(intermediateCounts) - 1; j >= 0; j-- {
			cur := intermediateCounts[j]
			if cur != smoothed[len(smoothed)-1].count && cur != tc[i].count {
				t := (dayPeriod + tc[i].time - time.Duration(j+1)*time.Minute) % dayPeriod
				smoothed = append(smoothed, timeCount{t, cur})
			}
		}
		smoothed = append(smoothed, tc[i])
	}
	return smoothed[:len(smoothed)-1], nil
}

var posError = errors.New("could not find position")

func findPos(tc []timeCount, cur int, offset time.Duration) int {
	first := true
	for i := cur; i != cur || first; i = (i + 1) % len(tc) {
		if tc[i].time > offset {
			return i
		}
		first = false
	}
	return 0
}

func (s *Scaler) setCount(c int) {
	glog.Infof("scaling to %d replicas", c)
	rcList, err := client.ReplicationControllers(namespace).List(s.selector)
	if err != nil {
		glog.Errorf("could not get replication controllers: %v", err)
		return
	}
	for _, rc := range rcList.Items {
		rc.Spec.Replicas = c
		if _, err = client.ReplicationControllers(namespace).Update(&rc); err != nil {
			glog.Errorf("unable to scale replication controller: %v", err)
		}
	}
}

func (s *Scaler) timeOffset() time.Duration {
	return time.Since(s.start) % dayPeriod
}

func (s *Scaler) curpos(offset time.Duration) int {
	return findPos(s.timeCounts, s.pos, offset)
}

func (s *Scaler) scale() {
	for {
		select {
		case <-s.done:
			return
		default:
			offset := s.timeOffset()
			s.pos = s.curpos(offset)
			if s.timeCounts[s.pos].time < offset {
				time.Sleep(dayPeriod - offset)
				continue
			}
			time.Sleep(s.timeCounts[s.pos].time - offset)
			s.setCount(s.timeCounts[s.pos].count)
		}
	}
}

func (s *Scaler) Start() error {
	now := time.Now().UTC()
	s.start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if *startNow {
		s.start = now
	}

	// set initial count
	pos := s.curpos(s.timeOffset())
	// add the len to avoid getting a negative index
	pos = (pos - 1 + len(s.timeCounts)) % len(s.timeCounts)
	s.setCount(s.timeCounts[pos].count)

	s.done = make(chan struct{})
	go s.scale()
	return nil
}

func safeclose(c chan<- struct{}) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()
	close(c)
	return nil
}

func (s *Scaler) Stop() error {
	if err := safeclose(s.done); err != nil {
		return errors.New("already stopped scaling")
	}
	return nil
}

var (
	counts       = flag.String("counts", "", "replica counts, must have at least one (csv)")
	times        = flag.String("times", "", "times to set replica counts relative to UTC following ISO 8601 (csv)")
	userLabels   = flag.String("labels", "", "replication controller labels, syntax should follow https://godoc.org/github.com/GoogleCloudPlatform/kubernetes/pkg/labels#Parse")
	maxScaleRate = flag.Float64("maxrate", 5, "max number of replicas that may be added/removed per minute")
	startNow     = flag.Bool("now", false, "times are relative to now not 0:00 UTC (for demos)")
	local        = flag.Bool("local", false, "set to true if running on local machine not within cluster")
	localPort    = flag.Int("localport", 8001, "port that kubectl proxy is running on (local must be true)")

	namespace string = os.Getenv("POD_NAMESPACE")

	client *kclient.Client
)

const usageNotes = `
counts and times must both be set and be of equal length. Example usage:
  diurnal -labels name=redis-slave -times 00:00,06:00 -counts 3,9 -maxrate 1
`

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
	fmt.Fprint(os.Stderr, usageNotes)
}

func main() {
	flag.Usage = usage
	flag.Parse()

	var (
		cfg *kclient.Config
		err error
	)
	if *local {
		cfg = &kclient.Config{Host: fmt.Sprintf("http://localhost:%d", *localPort)}
	} else {
		cfg, err = kclient.InClusterConfig()
		if err != nil {
			glog.Errorf("failed to load config: %v", err)
			flag.Usage()
			os.Exit(1)
		}
	}
	client, err = kclient.New(cfg)

	selector, err := labels.Parse(*userLabels)
	if err != nil {
		glog.Fatal(err)
	}
	tc, err := parseTimeCounts(*times, *counts)
	if err != nil {
		glog.Fatal(err)
	}
	if namespace == "" {
		glog.Fatal("POD_NAMESPACE is not set. Set to the namespace of the replication controller if running locally.")
	}
	scaler, err := NewScaler(tc, selector)
	if err != nil {
		glog.Fatal(err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM)

	glog.Info("starting scaling")
	if err := scaler.Start(); err != nil {
		glog.Fatal(err)
	}
	<-sigChan
	glog.Info("stopping scaling")
	if err := scaler.Stop(); err != nil {
		glog.Fatal(err)
	}
}
