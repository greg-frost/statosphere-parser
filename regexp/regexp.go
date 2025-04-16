package regexp

import (
	"fmt"
	"regexp"
	"sort"
	"sync"
	"time"
)

// Группа регулярных выражений
type group struct {
	group   string
	pattern string
	regexp  *regexp.Regexp
	pockets map[string]int
}

// Набор групп регулярных выражений
type groups struct {
	g map[string]*group
	t map[string][]time.Duration
	m sync.RWMutex
}

// Глобальное хранилище регулярных выражений
var rs = groups{g: make(map[string]*group), t: make(map[string][]time.Duration)}

// Запись в хранилище
func (rs *groups) setGroup(groupname string, r group) {
	rs.m.Lock()
	defer rs.m.Unlock()

	rs.g[groupname] = &r
}

// Чтение из хранилища
func (rs *groups) group(groupname string) group {
	rs.m.RLock()
	defer rs.m.RUnlock()

	r, ok := rs.g[groupname]
	if !ok {
		return group{}
	}

	return *r
}

// Очистка хранилища
func (rs *groups) clear() {
	*rs = groups{g: make(map[string]*group), t: make(map[string][]time.Duration)}
}

// Подготовка регулярного выражения к работе
func Prepare(group, pattern string) group {
	reg := rs.group(group)

	if reg.regexp == nil || reg.pattern != pattern {
		reg.group = group
		reg.pattern = pattern
		reg.regexp = regexp.MustCompile(reg.pattern)
		reg.pockets = map[string]int{"": 0}

		names := reg.regexp.SubexpNames()
		for i, n := range names {
			if n != "" {
				reg.pockets[n] = i
			}
		}

		rs.setGroup(group, reg)
	}

	return reg
}

// Поиск соответствия
func (r *group) Match(text string) bool {
	var res bool

	if text == "" {
		return res
	}

	if r.pattern == "" {
		return res
	}

	start := time.Now()

	res = r.regexp.MatchString(text)

	if isProfiler {
		profilerTime(r.group, time.Since(start))
	}

	return res
}

// Поиск подстроки
func (r *group) Find(text string) map[string]string {
	res := make(map[string]string)

	if text == "" {
		return res
	}

	if r.pattern == "" {
		return res
	}

	start := time.Now()

	found := r.regexp.FindStringSubmatch(text)
	if len(found) > 0 {
		for p, i := range r.pockets {
			res[p] = found[i]
		}
	}

	if isProfiler {
		profilerTime(r.group, time.Since(start))
	}

	return res
}

// Поиск всех подстрок
func (r *group) FindAll(text string) []map[string]string {
	res := make([]map[string]string, 0, 10)

	if text == "" {
		return res
	}

	if r.pattern == "" {
		return res
	}

	start := time.Now()

	found := r.regexp.FindAllStringSubmatch(text, -1)
	if len(found) > 0 {
		for f := range found {
			m := make(map[string]string)
			for p, i := range r.pockets {
				m[p] = found[f][i]
			}
			res = append(res, m)
		}
	}

	if isProfiler {
		profilerTime(r.group, time.Since(start))
	}

	return res
}

// Замена всех совпадений
func (r *group) ReplaceAll(text, replace string) string {
	var res string

	if text == "" {
		return res
	}

	if r.pattern == "" {
		return text
	}

	start := time.Now()

	res = r.regexp.ReplaceAllString(text, replace)

	if isProfiler {
		profilerTime(r.group, time.Since(start))
	}

	return res
}

/* REGEXP PROFILER */

// Профайлер времени выполнения
type profiler struct {
	group string
	min   time.Duration
	max   time.Duration
	avg   time.Duration
	sum   time.Duration
	count int
}

// Включение (выключение) профайлера
var isProfiler = true

// Запись времени выполнения
func profilerTime(group string, t time.Duration) {
	rs.m.Lock()
	defer rs.m.Unlock()

	rs.t[group] = append(rs.t[group], t)
}

// Подсчет времени выполнения
func ProfilerCount(minTime time.Duration) (total time.Duration, marks []profiler) {
	if len(rs.t) == 0 {
		return
	}

	rs.m.RLock()
	defer rs.m.RUnlock()

	for g, ts := range rs.t {
		var mark profiler

		mark.group = g
		mark.min = ts[0]
		mark.count = len(ts)

		for _, t := range ts {
			if t < mark.min {
				mark.min = t
			}
			if t > mark.max {
				mark.max = t
			}
			mark.sum += t
		}

		mark.avg = mark.sum / time.Duration(mark.count)
		total += mark.sum

		if mark.sum >= minTime {
			marks = append(marks, mark)
		}
	}

	return total, marks
}

// Печать отчета времени выполнения
func ProfilerReport(minTime time.Duration) {
	total, marks := ProfilerCount(minTime)

	if len(marks) == 0 {
		return
	}

	fmt.Print("Regexps\n-------\n")

	sort.Slice(marks, func(i, j int) bool {
		return marks[i].sum > marks[j].sum
	})

	for _, mark := range marks {
		fmt.Printf("%-18v %-8v %-12v [ %v ]\n", mark.group, mark.count, mark.max, mark.sum)
	}

	fmt.Printf("\nОбщее время выполнения: %v\n\n", total)
}
