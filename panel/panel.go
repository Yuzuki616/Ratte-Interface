package panel

import (
	"github.com/Yuzuki616/Ratte-Interface/baseplugin"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"net/rpc"
	"os/exec"
)

const PluginName = "ratte-panel"

var HandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "is-ratte-panel",
	MagicCookieValue: "true",
}

type Panel interface {
	AddRemote(params *AddRemoteParams) *AddRemoteRsp
	DelRemote(id int) error
	GetNodeInfo(id int) *GetNodeInfoRsp
	GetUserList(id int) *GetUserListRsp
	ReportUserTraffic(p *ReportUserTrafficParams) error
}

type PluginClient struct {
	Panel
	*baseplugin.Client
}

func (c *PluginClient) Close() error {
	return nil
}

func NewClient(l hclog.Logger, cmd *exec.Cmd) (client *PluginClient, err error) {
	pc, obj, err := baseplugin.NewClient(PluginName, cmd, l, NewPlugin(nil), HandshakeConfig)
	if err != nil {
		return nil, err
	}
	return &PluginClient{
		Client: pc,
		Panel:  obj.(Panel),
	}, nil
}

type PluginServer struct {
	*baseplugin.Server
}

func NewServer(l hclog.Logger, p Panel) (*PluginServer, error) {
	s, err := baseplugin.NewServer(PluginName, l, HandshakeConfig, NewPlugin(p))
	if err != nil {
		return nil, err
	}
	return &PluginServer{Server: s}, nil
}

type Plugin struct {
	p Panel
}

func NewPlugin(impl Panel) *Plugin {
	return &Plugin{
		p: impl,
	}
}

func (p *Plugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &PluginImplServer{p: p.p}, nil
}

func (_ *Plugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &PluginImplClient{c: c}, nil
}

type PluginImplServer struct {
	p Panel
}

var _ Panel = (*PluginImplClient)(nil)

type PluginImplClient struct {
	c *rpc.Client
}

func (c *PluginImplClient) call(method string, args interface{}, reply interface{}) error {
	return c.c.Call("Plugin."+method, args, reply)
}
