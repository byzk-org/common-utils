package plugintemplate

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"github.com/byzk-org/common-utils/hash"
	"github.com/byzk-org/common-utils/plugintemplate/telnet"
	"github.com/byzk-org/common-utils/plugintemplate/timelistener"
	"github.com/tjfoc/gmsm/sm2"
	"time"
)

func NewPluginTemplate(priKey *sm2.PrivateKey, goos, goarch, saveDir string) *pluginTemplate {
	return &pluginTemplate{
		priKey:  priKey,
		result:  make([]*PluginResult, 0),
		goos:    goos,
		goarch:  goarch,
		saveDir: saveDir,
	}
}

type PluginType string

const (
	PluginTypeListener PluginType = "listener"
	PluginTypeNormal   PluginType = "normal"
	PluginTypeBefore   PluginType = "before"
	PluginTypeAfter    PluginType = "after"
)

type PluginInfo struct {
	Type      PluginType    `json:"type,omitempty"`
	Md5       []byte        `json:"md5,omitempty"`
	Sha1      []byte        `json:"sha1,omitempty"`
	Sign      []byte        `json:"sign,omitempty"`
	EnvConfig []interface{} `json:"envConfig,omitempty"`
}

func (p *PluginInfo) Src() []byte {
	return bytes.Join([][]byte{
		[]byte(p.Type),
		p.Md5,
		p.Sha1,
	}, nil)
}

func (p *PluginInfo) String() string {
	marshal, _ := json.Marshal(p)
	return string(marshal)
}

type PluginResult struct {
	PluginInfo *PluginInfo
	Content    string
}

type pluginTemplate struct {
	result  []*PluginResult
	priKey  *sm2.PrivateKey
	goos    string
	goarch  string
	saveDir string
	Err     error
}

func (p *pluginTemplate) TimePlugin(expire time.Time) *pluginTemplate {
	if p.Err != nil {
		return p
	}
	timePlugin, err := timelistener.Create(expire, p.goos, p.goarch, p.saveDir)
	if err != nil {
		p.Err = err
		return p
	}

	p.convertData2Result(timePlugin, PluginTypeListener, nil)
	return p
}

func (p *pluginTemplate) TelnetBeforePlugin(host string, timeOut int, timeSpace int) *pluginTemplate {
	if p.Err != nil {
		return p
	}
	telnetPlugin, err, envConfig := telnet.CreateBefore(host, timeOut, timeSpace, p.goos, p.goarch, p.saveDir)
	if err != nil {
		p.Err = err
		return p
	}

	p.convertData2Result(telnetPlugin, PluginTypeBefore, envConfig)
	return p
}

func (p *pluginTemplate) TelnetAfterPlugin(host string, timeOut int, timeSpace int) *pluginTemplate {
	if p.Err != nil {
		return p
	}
	telnetPlugin, err, envConfig := telnet.CreateAfter(host, timeOut, timeSpace, p.goos, p.goarch, p.saveDir)
	if err != nil {
		p.Err = err
		return p
	}

	p.convertData2Result(telnetPlugin, PluginTypeAfter, envConfig)
	return p
}

func (p *pluginTemplate) Result() ([]*PluginResult, error) {
	return p.result, p.Err
}

func (p *pluginTemplate) convertData2Result(content string, pluginType PluginType, envConfig []interface{}) {
	md5Sum, err := hash.CalcMd5(content)
	if err != nil {
		p.Err = errors.New("计算插件MD5摘要失败")
		return
	}
	sha1Sum, err := hash.CalcSha1(content)
	if err != nil {
		p.Err = errors.New("计算插件SHA1摘要失败")
		return
	}

	pluginInfo := &PluginInfo{
		Type:      pluginType,
		Md5:       md5Sum[:],
		Sha1:      sha1Sum[:],
		EnvConfig: envConfig,
	}
	sign, err := p.priKey.Sign(rand.Reader, pluginInfo.Src(), nil)
	if err != nil {
		p.Err = errors.New("创建时间插件-创建签名失败")
		return
	}
	pluginInfo.Sign = sign
	p.result = append(p.result, &PluginResult{
		PluginInfo: pluginInfo,
		Content:    content,
	})
}
