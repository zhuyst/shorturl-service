package logger

import (
	"fmt"
	"testing"
)

type testLogger struct {
	msg string
}

const infoPrefix = "[INFO] "

func (t *testLogger) Info(format string, v ...interface{}) {
	t.msg = fmt.Sprintf(infoPrefix+format, v...)
}

func (t *testLogger) Error(format string, v ...interface{}) {
	t.msg = fmt.Sprintf("[ERROR] "+format, v...)
}

func (t *testLogger) Fatal(format string, v ...interface{}) {
	t.msg = fmt.Sprintf("[ERROR] "+format, v...)
}

func TestSetLogger(t *testing.T) {
	l := &testLogger{}
	Logger = l

	testMsg := "Testing"
	Logger.Info(testMsg)
	if l.msg != infoPrefix+testMsg {
		t.Errorf("SetLogger ERROR, expected l.msg == infoPrefix + testMsg, "+
			"got %s != %s", l.msg, infoPrefix+testMsg)
		return
	}
	t.Logf("SetLogger PASS")
}
