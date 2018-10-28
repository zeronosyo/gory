package log

import (
	"fmt"
	"log/syslog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

func InitLogger() *logrus.Logger {
	logger = logrus.New()
	logger.SetFormatter(&GoryFormatter{
		pn:  filepath.Base(os.Args[0]),
		pid: os.Getpid(),
	})
	if gin.Mode() == gin.ReleaseMode {
		logger.SetLevel(logrus.InfoLevel)
		writer, err := syslog.New(syslog.LOG_EMERG, "gory")
		if err != nil {
			logger.Fatalf("Init logger got error => %v", err)
		}
		logger.SetOutput(writer)
	} else {
		logger.SetLevel(logrus.DebugLevel)
		logger.SetOutput(os.Stdout)
	}
	return logger
}

func GetLogger() *logrus.Logger {
	return logger
}

type GoryFormatter struct {
	pn  string
	pid int
}

func (f *GoryFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	rlt := make([]string, 0)
	// RequestTime LogLevel ProcessName[PID]:
	rlt = append(rlt, fmt.Sprintf(
		"%v %v %v[%v]:",
		entry.Time.Format(time.RFC3339),
		entry.Level,
		f.pn,
		f.pid,
	))
	// [RequestIp - StatusCode Method Path RequestId]
	reqInfo := make([]string, 0)
	if reqIp, exist := entry.Data["RequestIp"]; exist {
		reqInfo = append(reqInfo, fmt.Sprint(reqIp))
		reqInfo = append(reqInfo, "-")
	}
	if status, exist := entry.Data["StatusCode"]; exist {
		reqInfo = append(reqInfo, fmt.Sprint(status))
	}
	if method, exist := entry.Data["RequestMethod"]; exist {
		reqInfo = append(reqInfo, fmt.Sprint(method))
	}
	if uri, exist := entry.Data["RequestURI"]; exist {
		reqInfo = append(reqInfo, fmt.Sprint(uri))
	}
	if reqId, exist := entry.Data["RequestId"]; exist {
		reqInfo = append(reqInfo, fmt.Sprint(reqId))
	}
	if len(reqInfo) > 0 {
		rlt = append(rlt, fmt.Sprintf("%v", reqInfo))
	}
	// log metas if exists: Key1 => Value1 Key2 => Value2
	if metas, exist := entry.Data["metas"]; exist {
		for k, v := range metas.(map[string]interface{}) {
			rlt = append(rlt, fmt.Sprintf("%v => %v", k, v))
		}
	}
	rlt = append(rlt, entry.Message)
	if cost, exist := entry.Data["Cost"]; exist {
		rlt = append(rlt, fmt.Sprintf("%vms", cost))
	}
	rlt = append(rlt, "\n")
	return []byte(strings.Join(rlt, " ")), nil
}

func LoggerMiddlerware(l *logrus.Logger) gin.HandlerFunc {
	if l == nil {
		l = logger
	}
	return func(c *gin.Context) {
		t := time.Now()
		requestId, err := uuid.NewV4()
		if err != nil {
			l.Errorf("Generate request id got error => %v", err)
		}
		loggerWithCtx := l.WithTime(t).WithFields(logrus.Fields{
			"RequestId":     requestId,
			"RequestIp":     c.ClientIP(),
			"RequestMethod": c.Request.Method,
			"RequestURI":    c.Request.RequestURI,
		})
		c.Set("logger", loggerWithCtx)
		logCtx := make(map[string]interface{})
		c.Set("_logCtx", logCtx)

		// process request
		c.Next()

		loggerWithCtx = loggerWithCtx.WithField("StatusCode", c.Writer.Status())
		if metas, exist := logCtx["metas"]; exist {
			loggerWithCtx = loggerWithCtx.WithField("metas", metas)
		}
		loggerWithCtx = loggerWithCtx.WithField(
			"Cost", float64(time.Now().Sub(t).Nanoseconds())/float64(time.Millisecond),
		)
		argInfo := make([]string, 0)
		if args, exist := logCtx["args"]; exist {
			for k, v := range args.(map[string]interface{}) {
				argInfo = append(argInfo, fmt.Sprintf("%v=%#v", k, v))
			}
		}
		// TODO log error or warning or info
		loggerWithCtx.Info(fmt.Sprintf(
			"%v(%v)", filepath.Base(c.HandlerName()), strings.Join(argInfo, ","),
		))
	}
}

func AddLogArgs(c *gin.Context, key string, value interface{}) {
	logCtx := c.MustGet("_logCtx").(map[string]interface{})
	args, exist := logCtx["args"]
	if !exist {
		args = make(map[string]interface{})
		logCtx["args"] = args
	}
	args.(map[string]interface{})[key] = value
}

func AddLogMeta(c *gin.Context, key string, value interface{}) {
	logCtx := c.MustGet("_logCtx").(map[string]interface{})
	metas, exist := logCtx["metas"]
	if !exist {
		metas = make(map[string]interface{})
		logCtx["metas"] = metas
	}
	metas.(map[string]interface{})[key] = value
}
