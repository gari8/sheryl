package types

import (
	"bytes"
	"context"
	"github.com/gari8/sheryl/pkg/utils"
	"os/exec"
	"reflect"
	"strings"
	"text/template"
	"time"
)

type Step struct {
	Name     string
	Env      []string
	Cmd      string
	Delay    time.Duration
	Retries  int
	Interval time.Duration
	Output   string
	StartAt  time.Time
	EndAt    time.Time
	Duration time.Duration
	PID      int
	Verbose  bool
	Failed   bool
	Err      error
}

type StepOutput struct {
	Name     string        `json:"name,omitempty"`
	Env      []string      `json:"env,omitempty"`
	Cmd      string        `json:"cmd,omitempty"`
	Delay    time.Duration `json:"delay,omitempty"`
	Retries  int           `json:"retries,omitempty"`
	Interval time.Duration `json:"interval,omitempty"`
	Output   string        `json:"output,omitempty"`
	StartAt  time.Time     `json:"startAt,omitempty"`
	EndAt    time.Time     `json:"endAt,omitempty"`
	Duration time.Duration `json:"duration,omitempty"`
	PID      int           `json:"pid,omitempty"`
}

func (s *Step) ToStepOutput() *StepOutput {
	return &StepOutput{
		Name:     s.Name,
		Env:      s.Env,
		Cmd:      s.Cmd,
		Delay:    s.Delay,
		Retries:  s.Retries,
		Interval: s.Interval,
		Output:   s.Output,
		StartAt:  s.StartAt,
		EndAt:    s.EndAt,
		Duration: s.Duration,
		PID:      s.PID,
	}
}

type SimpleStepOutput struct {
	Name     string        `json:"name,omitempty"`
	Env      []string      `json:"env,omitempty"`
	Cmd      string        `json:"cmd,omitempty"`
	Output   string        `json:"output,omitempty"`
	Duration time.Duration `json:"duration,omitempty"`
}

func (s *Step) ToSimpleStepOutput() *SimpleStepOutput {
	return &SimpleStepOutput{
		Name:     s.Name,
		Env:      s.Env,
		Cmd:      s.Cmd,
		Output:   s.Output,
		Duration: s.Duration,
	}
}

func (s *Step) Run(ctx context.Context, beforeSteps map[string]*Step) ([]byte, error) {
	time.Sleep(s.Delay)
	parsed, err := template.New(s.Name).Parse(s.Cmd)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err = parsed.Execute(&buf, toLowerKeyMap(beforeSteps))
	if err != nil {
		return nil, err
	}
	cmd := exec.CommandContext(ctx, "sh", "-c", buf.String())
	cmd.Env = s.Env
	startAt := time.Now()
	output, err := cmd.CombinedOutput()
	endAt := time.Now()
	duration := endAt.Sub(startAt)
	s.Cmd = buf.String()
	s.StartAt = startAt
	s.EndAt = endAt
	s.Duration = duration
	s.PID = cmd.Process.Pid
	// TODO: 測定項目を追加する
	//if sysUsage, ok := cmd.ProcessState.SysUsage().(*syscall.Rusage); ok {
	//
	//}
	if s.Failed = err != nil; s.Failed {
		return nil, err
	}
	return output, nil
}

type convertibleToMap interface {
	toBeMap() map[string]any
}

func toLowerKeyMap[T convertibleToMap](oldMap map[string]T) map[string]any {
	newMap := make(map[string]any)
	for k, v := range oldMap {
		newMap[strings.ToLower(k)] = v.toBeMap()
	}
	return newMap
}

func (s *Step) toBeMap() map[string]any {
	result := make(map[string]any)
	val := reflect.ValueOf(s)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	typ := val.Type()

	if val.Kind() == reflect.Struct {
		for i := 0; i < val.NumField(); i++ {
			field := typ.Field(i)
			fieldName, _ := utils.GetJsonTag(field)
			if fieldName != "" {
				result[fieldName] = val.Field(i).Interface()
			}
		}
	}
	return result
}
