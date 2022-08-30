package conf

import (
	"context"
	"net/url"

	"github.com/apolloconfig/agollo/v4"
	"github.com/apolloconfig/agollo/v4/env/config"
	"github.com/apolloconfig/agollo/v4/storage"
	"github.com/gotomicro/ego/core/econf"
	"github.com/gotomicro/ego/core/econf/manager"
	"github.com/gotomicro/ego/core/elog"
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

	c := &config.AppConfig{
		AppID:          urlInfo.Query().Get("appId"),
		Cluster:        urlInfo.Query().Get("cluster"),
		IP:             urlInfo.Host,
		NamespaceName:  fp.namespace,
		IsBackupConfig: true,
		Secret:         urlInfo.Query().Get("secret"),
	}
	agollo.SetLogger(fp.logger.ZapSugaredLogger())
	client, err := agollo.StartWithConfig(func() (*config.AppConfig, error) {
		return c, nil
	})
	if err != nil {
		fp.logger.Panic("agollo start fail", elog.FieldErr(err))
	}
	fp.apollo = client
	fp.key = configKey
	fp.enableWatch = watch

	if watch {
		fp.changed = make(chan struct{}, 1)
		//fp.apollo.OnUpdate(func(event *agollo.ChangeEvent) {
		//	fp.changed <- struct{}{}
		//})
		fp.apollo.AddChangeListener(fp)
	}
	return econf.ConfigType(configType)
}

// ReadConfig ...
func (fp *dataSource) ReadConfig() (content []byte, err error) {
	//value := fp.apollo.GetString(fp.key, agollo.WithNamespace(fp.namespace))
	value := fp.apollo.GetConfig(fp.namespace).GetValue(fp.key)
	return []byte(value), nil
}

// Close ...
func (fp *dataSource) Close() error {
	close(fp.changed)
	//return fp.apollo.Close()
	return nil
}

// IsConfigChanged ...
func (fp *dataSource) IsConfigChanged() <-chan struct{} {
	return fp.changed
}

//OnChange 增加变更监控
func (fp *dataSource) OnChange(event *storage.ChangeEvent) {
	fp.changed <- struct{}{}
}

//OnNewestChange 监控最新变更
func (fp *dataSource) OnNewestChange(event *storage.FullChangeEvent) {

}
