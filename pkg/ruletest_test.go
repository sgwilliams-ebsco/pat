package pkg

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/prometheus/prometheus/rules"
	"time"
)

func TestGetTestCaseName(t *testing.T) {
	prt := PromRuleTest{
		Name: "Test HTTP Requests too low alert",
		Assertions: []Assertion{
			{
				At: "5m",
			},
			{
				At: "10m",
			},
		},
	}
	assert.Equal(t, "Test_HTTP_Requests_too_low_alert_at_5m", prt.getTestCaseName(0))
	assert.Equal(t, "Test_HTTP_Requests_too_low_alert_at_10m", prt.getTestCaseName(1))
}

func TestNewPromRuleTestFromFile(t *testing.T) {
	promRuleTest, err := NewPromRuleTestFromFile("testdata/test.yaml")
	assert.Nil(t, err)
	assert.Equal(t, "Test HTTP Requests too low alert", promRuleTest.Name)
	assert.Equal(t, "rules.yaml", promRuleTest.Rules.FromFile)
	assert.Equal(t, 2, len(promRuleTest.Fixtures[0].Metrics))
	assert.Equal(t, 3, len(promRuleTest.Assertions))
	assert.Equal(t, []Alert{}, promRuleTest.Assertions[2].Expected)
	assert.Equal(t, "testdata/test.yaml", promRuleTest.filename)
}

func TestNewPromRuleTestFromString(t *testing.T) {
	fileContent, err := ioutil.ReadFile("testdata/test.yaml")
	assert.Nil(t, err)

	promRuleTest, err := NewPromRuleTestFromString(fileContent)
	assert.Nil(t, err)

	assert.Equal(t, "Test HTTP Requests too low alert", promRuleTest.Name)
	assert.Equal(t, "rules.yaml", promRuleTest.Rules.FromFile)
	assert.Equal(t, 2, len(promRuleTest.Fixtures[0].Metrics))
	assert.Equal(t, 3, len(promRuleTest.Assertions))
	assert.Equal(t, []Alert{}, promRuleTest.Assertions[2].Expected)
	assert.Equal(t, FilenameInline, promRuleTest.filename)
}

func TestAlertLessThan(t *testing.T) {
	alerts := []map[string]string{
		{"__name__": "ALERTS", "alertname": "FOO", "alertstate": "pending", "instance": "1"},
		{"__name__": "ALERTS", "alertname": "FOO", "alertstate": "pending", "instance": "0"},
	}
	assert.False(t, alertLessThan(alerts[0], alerts[1]))
	assert.True(t, alertLessThan(alerts[1], alerts[0]))

	alerts = []Alert{
		{"__name__": "ALERTS", "alertname": "FOO", "alertstate": "pending", "instance": "0"},
		{"__name__": "ALERTS", "alertname": "FOO", "alertstate": "pending", "instance": "1"},
	}

	assert.False(t, alertLessThan(alerts[1], alerts[0]))
	assert.True(t, alertLessThan(alerts[0], alerts[1]))
}

func TestAssertAlertsEqual(t *testing.T) {
	testCases := []struct{
		expected []Alert
		actual   []map[string]string
		comment  string
	}{
		{
			expected: []Alert{
				{"alertname": "FOO", "alertstate": "pending"},
			},
			actual: []map[string]string{
				{"__name__": "ALERTS", "alertname": "FOO", "alertstate": "pending"},
			},
			comment: "Single alert should work",
		},
		{
			expected: []Alert{
				{"alertname": "FOO", "alertstate": "pending", "instance": "0"},
				{"alertname": "FOO", "alertstate": "pending", "instance": "1"},
			},
			actual: []map[string]string{
				{"__name__": "ALERTS", "alertname": "FOO", "alertstate": "pending", "instance": "0"},
				{"__name__": "ALERTS", "alertname": "FOO", "alertstate": "pending", "instance": "1"},
			},
			comment: "Multiple alerts of the same order should work",
		},
		{
			expected: []Alert{
				{"alertname": "FOO", "alertstate": "pending", "instance": "0"},
				{"alertname": "FOO", "alertstate": "pending", "instance": "1"},
			},
			actual: []map[string]string{
				{"__name__": "ALERTS", "alertname": "FOO", "alertstate": "pending", "instance": "1"},
				{"__name__": "ALERTS", "alertname": "FOO", "alertstate": "pending", "instance": "0"},
			},
			comment: "Multiple alerts of the different order should also work",
		},
		{
			expected: []Alert{
				{"foo": "bar", "__name__": "superalert"},
			},
			actual: []map[string]string{
				{"foo": "bar", "__name__": "superalert"},
			},
			comment: "If __name__ is specified by the user, do not override it",
		},
	}

	for _, tc := range testCases {
		assert.True(t, assertAlertsEqual(t, tc.expected, tc.actual), tc.comment)
	}
}

func TestEvalRuleGroupAtInstantWithNoAlert(t *testing.T) {
	prt := PromRuleTest{}
	noAlerts, err := prt.evalRuleGroupAtInstant(nil, []*rules.Group{}, time.Now())

	assert.Nil(t, err)
	assert.Equal(t, []Alert{}, noAlerts)
}
