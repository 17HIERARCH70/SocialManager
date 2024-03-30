package logger

import (
	"github.com/sirupsen/logrus"
	"os"
)

var Log = logrus.New()

func init() {
	Log.Out = os.Stdout
	Log.SetLevel(logrus.DebugLevel) // Or any desired log level
}
