package regexp

import (
	"reflect"
	"regexp"
	"statosphere/parser/mock"
	"strings"
	"testing"
	"time"
)

func TestSetGroup(t *testing.T) {
	rs.clear()

	tests := []struct {
		test   string
		value  string
		result group
	}{
		{"Valid", "group", group{group: "group", pattern: `test`}},
		{"Empty", "", group{}},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			rs.setGroup(tt.value, tt.result)
			result := *rs.g[tt.value]

			if !reflect.DeepEqual(result, tt.result) {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestGroup(t *testing.T) {
	rs.clear()

	tests := []struct {
		test   string
		value  string
		isSet  bool
		result group
	}{
		{"Valid", "group", true, group{group: "group", pattern: `test`}},
		{"NotSet", "another", false, group{}},
		{"Empty", "", true, group{}},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			if tt.isSet {
				rs.g[tt.value] = &tt.result
			}
			result := rs.group(tt.value)

			if !reflect.DeepEqual(result, tt.result) {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestClear(t *testing.T) {
	rs.clear()
	rs.setGroup("test", group{})

	tests := []struct {
		test    string
		isClear bool
		result  int
	}{
		{"None", false, 1},
		{"Clear", true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			if tt.isClear {
				rs.clear()
			}
			result := len(rs.g)

			if result != tt.result {
				t.Errorf("Получено количество: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestPrepare(t *testing.T) {
	rs.clear()

	pockets := map[string]int{"": 0, "number": 1, "string": 3}
	noPockets := map[string]int{"": 0}

	tests := []struct {
		test    string
		value   string
		pattern string
		pockets map[string]int
		result  group
		count   int
	}{
		{"Valid", "group", `test`, noPockets, group{group: "group"}, 1},
		{"Copy", "group", `test`, noPockets, group{group: "group"}, 1},
		{"Diff", "group", `(?is)diff`, noPockets, group{group: "group"}, 1},
		{"Another", "another", `\d+`, noPockets, group{group: "another"}, 2},
		{"Pockets", "pockets", `(?P<number>\d+)(.*?)(?P<string>[A-Za-z]+)`,
			pockets, group{group: "pockets"}, 3},
		{"Empty", "", "", noPockets, group{}, 4},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			tt.result.pattern = tt.pattern
			tt.result.regexp = regexp.MustCompile(tt.pattern)
			tt.result.pockets = tt.pockets

			result := Prepare(tt.value, tt.pattern)
			count := len(rs.g)

			if !reflect.DeepEqual(result, tt.result) {
				t.Errorf("Получено значение: %#v, ожидается: %#v", result, tt.result)
			}
			if count != tt.count {
				t.Errorf("Получено количество: %v, ожидается: %v", count, tt.count)
			}
		})
	}
}

func TestMatch(t *testing.T) {
	rs.clear()
	isProfiler = true

	text := "Simple text \n with number 100"

	tests := []struct {
		test    string
		value   string
		pattern string
		result  bool
	}{
		{"Text", text, `text`, true},
		{"Number", text, `\d+`, true},
		{"Caseless", text, `(?i)simple`, true},
		{"NotFoundCase", text, `simple`, false},
		{"Multiline", text, `(?s)text.*?number`, true},
		{"NotFoundMulti", text, `text.*?number`, false},
		{"NoPattern", text, "", false},
		{"NoText", "", `text`, false},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			re := Prepare("match", tt.pattern)
			result := re.Match(tt.value)

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestFind(t *testing.T) {
	rs.clear()
	isProfiler = true

	text := "Simple text \n with number 1000"

	tests := []struct {
		test    string
		value   string
		pattern string
		pocket  string
		result  string
	}{
		{"Text", text, `text`, "", "text"},
		{"Mask", text, `t.{1,3}`, "", "text"},
		{"Number", text, `\d+`, "", "1000"},
		{"PocketText", text, `(?s)(?P<txt>t.{1,3}).*?(?P<num>\d+)`, "txt", "text"},
		{"PocketNumber", text, `(?s)(?P<txt>t.{1,3}).*?(?P<num>[0-9]+)`, "num", "1000"},
		{"NoPattern", text, "", "", ""},
		{"NoText", "", `text`, "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			re := Prepare("find", tt.pattern)
			results := re.Find(tt.value)
			result := results[tt.pocket]

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestFindAll(t *testing.T) {
	rs.clear()
	isProfiler = true

	text := "Simple text \n with number 1000"

	tests := []struct {
		test    string
		value   string
		pattern string
		pocket  string
		result  []string
	}{
		{"Text", text, `t.`, "", []string{"te", "t ", "th"}},
		{"Mask", text, `(?i)[site]{2,3}`, "", []string{"Si", "te", "it"}},
		{"Number", text, `\d`, "", []string{"1", "0", "0", "0"}},
		{"PocketText", text, `\b(?P<txt>\w{3,4})\b`, "txt", []string{"text", "with", "1000"}},
		{"PocketNumber", text, `(?P<num>[0-9]{2})`, "num", []string{"10", "00"}},
		{"NoPattern", text, "", "", []string{}},
		{"NoText", "", `text`, "", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			re := Prepare("findAll", tt.pattern)
			results := re.FindAll(tt.value)

			result := []string{}
			for _, res := range results {
				result = append(result, res[tt.pocket])
			}

			if !reflect.DeepEqual(result, tt.result) {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestReplaceAll(t *testing.T) {
	rs.clear()
	isProfiler = true

	text := "Simple text \n with number 1000"

	tests := []struct {
		test    string
		pattern string
		replace string
		result  string
	}{
		{"Text", `(text|number)`, "*", "Simple * \n with * 1000"},
		{"Newline", `\n `, "", "Simple * with * 1000"},
		{"Number", `\d+`, "{num}", "Simple * with * {num}"},
		{"Asterisk", `\* `, "", "Simple with {num}"},
		{"Denumber", `{.*?}`, "500", "Simple with 500"},
		{"NoPattern", "", "", "Simple with 500"},
		{"Clear", `.*`, "", ""},
		{"NoText", `text`, "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			re := Prepare("replaceAll", tt.pattern)
			result := re.ReplaceAll(text, tt.replace)
			text = result

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestProfilerTime(t *testing.T) {
	rs.clear()

	tests := []struct {
		test   string
		value  string
		result time.Duration
	}{
		{"Valid", "group", time.Millisecond},
		{"Again", "group", time.Second},
		{"NoTime", "another", 0},
		{"NoGroup", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			profilerTime(tt.value, tt.result)
			result := rs.t[tt.value][len(rs.t[tt.value])-1]

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
		})
	}
}

func TestProfilerCount(t *testing.T) {
	tests := []struct {
		test    string
		value   map[string][]int
		minTime int
		result  int
		count   int
	}{
		{"Fast", map[string][]int{"fast": {3, 5}}, 0, 8, 1},
		{"Slow", map[string][]int{"slow": {1670, 550, 2145}}, 0, 4365, 1},
		{"Both", map[string][]int{"fast": {3, 5}, "slow": {1670, 550, 2145}}, 0, 4373, 2},
		{"Filter", map[string][]int{"fast": {3, 5}, "slow": {1670, 550, 2145}}, 1000, 4373, 1},
		{"Empty", map[string][]int{}, 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			rs.clear()

			for group, times := range tt.value {
				for _, timeMs := range times {
					profilerTime(group, time.Duration(timeMs)*time.Millisecond)
				}
			}

			total, marks := ProfilerCount(time.Duration(tt.minTime) * time.Millisecond)
			result := int(total / time.Millisecond)
			count := len(marks)

			if result != tt.result {
				t.Errorf("Получено значение: %v, ожидается: %v", result, tt.result)
			}
			if count != tt.count {
				t.Errorf("Получено количество: %v, ожидается: %v", count, tt.count)
			}
		})
	}
}

func TestProfilerReport(t *testing.T) {
	rs.clear()

	profilerTime("fast", 3*time.Millisecond)
	profilerTime("fast", 5*time.Millisecond)
	profilerTime("slow", 550*time.Millisecond)
	profilerTime("slow", 1670*time.Millisecond)
	profilerTime("slow", 2145*time.Millisecond)

	tests := []struct {
		test    string
		minTime time.Duration
		result  []string
	}{
		{"Full", 0, []string{"Regexps", "---", "slow", "2.145s", "4.365s", "fast", "5ms", "8ms"}},
		{"Filter", 1000 * time.Millisecond, []string{"Regexps", "---", "slow", "2.145s", "4.365s"}},
		{"Empty", 10 * time.Second, []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			result := mock.ReadOutput(func() {
				ProfilerReport(tt.minTime)
			})

			for _, match := range tt.result {
				if !strings.Contains(result, match) {
					t.Fatalf("Получено значение: %q, ожидаются совпадения: %q, не найдено: %q",
						result, tt.result, match)
				}
			}
		})
	}
}
