package conf

import (
	"context"
	"net/url"

	"github.com/gotomicro/ego/core/econf"
	"github.com/gotomicro/ego/core/econf/manager"
	"github.com/gotomicro/ego/core/elog"
	"github.com/philchia/agollo/v4"
	"go.uber.org/zap"
)

// dataSource file provider.
type dataSource struct {
	key         string
	enableWatch bool
	namespace   string
	changed     chan struct{}
	cancel      context.CancelFunc
	logger      *elog.Component
	apollo      agollo.Client
}

func init() {
	manager.Register("apollo", &dataSource{})
}

// Parse 解析
func (fp *dataSource) Parse(path string, watch bool) econf.ConfigType {
	fp.logger = elog.EgoLogger.With(elog.FieldComponent(econf.PackageName))
	urlInfo, err := url.Parse(path)
	if err != nil {
		fp.logger.Panic("new datasource", elog.FieldErr(err))
		return ""
	}

	configKey := urlInfo.Query().Get("configKey")
	configType := urlInfo.Query().Get("configType")
	fp.namespace = urlInfo.Query().Get("namespaceName")

	if configKey == "" {
		fp.logger.Panic("key is empty")
	}

	if configType == "" {
		fp.logger.Panic("configType is empty")
	}

	apolloConf := agollo.Conf{
		AppID:              urlInfo.Query().Get("appId"),
		Cluster:            urlInfo.Query().Get("cluster"),
		NameSpaceNames:     []string{fp.namespace},
		MetaAddr:           urlInfo.Host,
		InsecureSkipVerify: true,
		AccesskeySecret:    urlInfo.Query().Get("accesskeySecret"),
		CacheDir:           ".",
	}

	fp.apollo = agollo.NewClient(&apolloConf, agollo.WithLogger(&agolloLogger{
		sugar: fp.logger.ZapSugaredLogger(),
	}))
	fp.key = configKey
	fp.enableWatch = watch
	err = fp.apollo.Start()
	if err != nil {
		fp.logger.Panic("agollo start fail", elog.FieldErr(err))
	}
	if watch {
		fp.changed = make(chan struct{}, 1)
		fp.apollo.OnUpdate(func(event *agollo.ChangeEvent) {
			fp.changed <- struct{}{}
		})
	}
	return econf.ConfigType(configType)
}

// ReadConfig ...
func (fp *dataSource) ReadConfig() (content []byte, err error) {
	value := fp.apollo.GetString(fp.key, agollo.WithNamespace(fp.namespace))
	return []byte(value), nil
}

// Close ...
func (fp *dataSource) Close() error {
	close(fp.changed)
	return fp.apollo.Stop()
}

// IsConfigChanged ...
func (fp *dataSource) IsConfigChanged() <-chan struct{} {
	return fp.changed
}

type agolloLogger struct {
	sugar *zap.SugaredLogger
}

// Infof ...
func (l *agolloLogger) Infof(format string, args ...interface{}) {
	l.sugar.Infof(format, args...)
}

// Errorf ...
func (l *agolloLogger) Errorf(format string, args ...interface{}) {
	l.sugar.Errorf(format, args...)
}
