package toast

import (
	"strings"
	"testing"
)

func TestBuildToastXMLBasics(t *testing.T) {
	xml := buildToastXML(toastContent{
		Title:               "Grok 运行完成",
		Message:             "学习ai编程项目/agent-notify\n任务已完成",
		IconURI:             "file:///C:/x/toast-icon.png",
		ActivationType:      "protocol",
		ActivationArguments: "anfocus:123",
		Audio:               "ms-winsoundevent:Notification.Default",
		Duration:            "long",
	})
	for _, want := range []string{
		`encoding="utf-8"`,
		`duration="long"`,
		`launch="anfocus:123"`,
		`<image placement="appLogoOverride" hint-crop="none" src="file:///C:/x/toast-icon.png"/>`,
		`<![CDATA[Grok 运行完成]]>`,
		`<![CDATA[学习ai编程项目/agent-notify]]>`,
		`<![CDATA[任务已完成]]>`,
		`<audio src="ms-winsoundevent:Notification.Default" loop="false"/>`,
	} {
		if !strings.Contains(xml, want) {
			t.Fatalf("missing %q in:\n%s", want, xml)
		}
	}
	if strings.Count(xml, "<text>") < 3 { // title + 2 body lines
		t.Fatalf("want >=3 text nodes: %s", xml)
	}
}

func TestBuildToastXMLNoIconDefaults(t *testing.T) {
	xml := buildToastXML(toastContent{Title: "t", Message: "b"})
	if strings.Contains(xml, "<image") {
		t.Fatalf("no image expected: %s", xml)
	}
	if !strings.Contains(xml, `duration="short"`) || !strings.Contains(xml, `<audio silent="true"/>`) {
		t.Fatalf("defaults wrong: %s", xml)
	}
}

func TestBuildToastXMLEscaping(t *testing.T) {
	xml := buildToastXML(toastContent{
		Title:               "a]]>b",
		ActivationType:      `proto"col`,
		ActivationArguments: `arg&1`,
	})
	if strings.Contains(xml, "a]]>b") {
		t.Fatalf("CDATA terminator not escaped: %s", xml)
	}
	if !strings.Contains(xml, "a]]]]><![CDATA[>b") {
		t.Fatalf("CDATA escape form missing: %s", xml)
	}
	if !strings.Contains(xml, `activationType="proto&quot;col"`) || !strings.Contains(xml, `launch="arg&amp;1"`) {
		t.Fatalf("attr escaping wrong: %s", xml)
	}
}

func TestBuildToastXMLActions(t *testing.T) {
	xml := buildToastXML(toastContent{
		Message: "m",
		Actions: []Action{{Type: "protocol", Label: "Open", Arguments: "bingmaps:?q=x"}},
	})
	if !strings.Contains(xml, `<action activationType="protocol" content="Open" arguments="bingmaps:?q=x"/>`) {
		t.Fatalf("action missing: %s", xml)
	}
}

func TestPowershellSingleQuote(t *testing.T) {
	if got := powershellSingleQuote(`a'b`); got != `a''b` {
		t.Fatalf("got %q", got)
	}
}
