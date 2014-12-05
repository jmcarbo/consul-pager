package consul_pager

import (
	"github.com/armon/consul-api"
	. "github.com/franela/goblin"
	"os"
	"os/exec"
	"testing"
	"time"
)

func initConsul() {
	os.RemoveAll("/tmp/consul")
	exec.Command("consul", "agent", "-server", "-bootstrap", "-data-dir=/tmp/consul").Start()
	time.Sleep(time.Second * 15)
}

func stopConsul() {
	exec.Command("killall", "-TERM", "consul").Run()
}

func TestConsulPager(t *testing.T) {

	g := Goblin(t, "-goblin.timeout=20s")

	g.Describe("pager", func() {
		g.Before(func() {
			initConsul()
		})

		g.After(func() {
			stopConsul()
		})

		g.It("can be instantiated", func() {
			client := Connect("", "", "")
			cp := NewConsulPager(client)
			g.Assert(cp.IsStopped()).IsTrue()
		})

		g.It("can be run and stopped", func() {
			client := Connect("", "", "")
			cp := NewConsulPager(client)
			cp.Run()
			time.Sleep(time.Second * 2)
			g.Assert(cp.IsLeader()).IsTrue()
			cp.Stop()
			g.Assert(cp.IsStopped()).IsTrue()
			cp.Destroy()
		})

		g.It("creates alert on failed check", func() {
			client := Connect("", "", "")
			client.Agent().CheckRegister(&consulapi.AgentCheckRegistration{"ACheck", "Acheck", "",
				consulapi.AgentServiceCheck{Interval: "2s", Script: "ping -c 1 loc"}})
			cp := NewConsulPager(client)
			cp.Run()
			time.Sleep(time.Second * 5)
			g.Assert(cp.HasAlert("Acheck")).IsTrue()
			cp.Stop()
			cp.Destroy()
		})

		g.It("creates alert and matures on failed check", func() {
			client := Connect("", "", "")
			client.Agent().CheckRegister(&consulapi.AgentCheckRegistration{"ACheck", "Acheck", "",
				consulapi.AgentServiceCheck{Interval: "2s", Script: "ping -c 1 loc"}})
			cp := NewConsulPager(client)
			cp.Run()
			g.Assert(cp.AlertIsMature("Acheck", "13s")).IsFalse()
			time.Sleep(time.Second * 8)
			g.Assert(cp.AlertIsMature("Acheck", "3s")).IsTrue()
			cp.Stop()
			cp.Destroy()
		})

		g.It("creates alert and matures on failed check and sends alert to channel", func() {
			client := Connect("", "", "")
			client.Agent().CheckRegister(&consulapi.AgentCheckRegistration{"ACheck", "Acheck", "",
				consulapi.AgentServiceCheck{Interval: "2s", Script: "ping -c 1 loc"}})
			cp := NewConsulPager(client)
			channel := NewSlackChannel("blabla", "blablalbalba", "#consul", "consul-pager")
			channel.SetId("MyChannel")
			cp.AddChannel(Channel{ChannelI: channel})
			cp.Run()
			g.Assert(cp.AlertIsMature("Acheck", "13s")).IsFalse()
			time.Sleep(time.Second * 8)
			g.Assert(cp.AlertIsMature("Acheck", "3s")).IsTrue()
			g.Assert(cp.AlertIsSended("Acheck")).IsTrue()
			cp.Stop()
			cp.Destroy()
		})

		g.It("It can be save to YAML", func() {
			client := Connect("", "", "")
			client.Agent().CheckRegister(&consulapi.AgentCheckRegistration{"ACheck", "Acheck", "",
				consulapi.AgentServiceCheck{Interval: "2s", Script: "ping -c 1 loc"}})
			cp := NewConsulPager(client)
			channel := NewSlackChannel("blabla", "blabla", "#consul", "consul-pager")
			channel.SetId("MyChannel")
			cp.AddChannel(Channel{ChannelI: channel})
			channel2 := NewSlackChannel("blabla", "balbal", "#consul", "consul-pager")
			channel2.SetId("MyChannel2")
			cp.AddChannel(Channel{ChannelI: channel2})
			cp.SaveToYAML("backup.yaml")
		})

		g.It("It can loaded from YAML", func() {
			client := Connect("", "", "")
			cp := LoadPagerFromYAML("backup.yaml", client)
			g.Assert(cp != nil).IsTrue()
		})
	})
}
