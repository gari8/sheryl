package types

import (
	"bytes"
	"cmp"
	"context"
	"log/slog"
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
	Name   string `json:"name,omitempty"`
	Cmd    string `json:"cmd,omitempty"`
	Output string `json:"output,omitempty"`
}

func (s *Step) ToSimpleStepOutput() *SimpleStepOutput {
	return &SimpleStepOutput{
		Name:   s.Name,
		Cmd:    s.Cmd,
		Output: s.Output,
	}
}

func (s *Step) Attributes(attrs ...any) []any {
	var val reflect.Value
	var typ reflect.Type
	if s.Verbose {
		val = reflect.ValueOf(s.ToStepOutput())
		typ = reflect.TypeOf(s.ToStepOutput())
	} else {
		val = reflect.ValueOf(s.ToSimpleStepOutput())
		typ = reflect.TypeOf(s.ToSimpleStepOutput())
	}

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldName, omitempty := getJsonTag(field)
		fieldValue := val.Field(i)
		if fieldValue.IsZero() && omitempty {
			continue
		}
		attrs = append(attrs, slog.Any(fieldName, fieldValue.Interface()))
	}
	return attrs
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
	if err != nil {
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
			fieldName, _ := getJsonTag(field)
			if fieldName != "" {
				result[fieldName] = val.Field(i).Interface()
			}
		}
	}
	return result
}

func getJsonTag(field reflect.StructField) (tag string, omitempty bool) {
	// json tag を取得
	var jsonTag string
	jsonTags := strings.Split(field.Tag.Get("json"), ",")
	if len(jsonTags) > 0 {
		jsonTag = jsonTags[0]
	}
	// omitempty があるか確認
	if len(jsonTags) > 1 {
		omitempty = jsonTags[1] == "omitempty"
	}
	return cmp.Or(jsonTag, strings.ToLower(field.Name)), omitempty
}
