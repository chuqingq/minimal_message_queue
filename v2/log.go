package mmq

import (
	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger = logrus.New()

func init() {
	logger.SetLevel(logrus.FatalLevel)
}
