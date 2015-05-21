package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/flier/curator.go"
)

// The action type; any of create/delete/setvalue.
type ZkActionType int

const (
	NONE   ZkActionType = iota
	CREATE              // creates <zknode> recussively
	DELETE              // deletes <zknode> recursively
	VALUE               // sets <value> to <zknode>
)

type ZkAction struct {
	Type     ZkActionType // action of this instance
	Key      string       // ZK node key
	NewValue string       // value to be set, if action is setvalue
	OldValue string       // existing value of the ZK node key
}

type ZkActions []*ZkAction

type ZkActionHandler interface {
	Handle(action *ZkAction) error
}

type ZkActionPrinter struct {
	Out *os.File
}

func (p *ZkActionPrinter) Handle(action *ZkAction) error {
	var buf bytes.Buffer

	switch action.Type {
	case CREATE:
		fmt.Fprintf(&buf, "CREATE- key: %s\n", action.Key)
	case DELETE:
		fmt.Fprintf(&buf, "DELETE- key: %s\n", action.Key)
	case VALUE:
		fmt.Fprintf(&buf, "VALUE- key: %s value: %s", action.Key, action.NewValue)

		if len(action.OldValue) > 0 {
			fmt.Fprintf(&buf, " old: %s", action.OldValue)
		}

		fmt.Fprintln(&buf)
	}

	fmt.Print(buf.String())

	return nil
}

type ZkActionExecutor struct{}

func (e *ZkActionExecutor) Handle(action *ZkAction) error {
	return nil
}

type ZkActionInteractiveExecutor struct{}

func (e *ZkActionInteractiveExecutor) Handle(action *ZkAction) error {
	return nil
}

type ZkNode struct {
	XMLName  xml.Name
	Name     string `xml:"name,attr,omitempty"`
	Value    string `xml:"value,attr,omitempty"`
	Ignore   *bool  `xml:"ignore,attr,omitempty"`
	Children []*ZkNode
}

type ZkTree interface {
	Dump(depth int) (string, error)
}

type ZkLiveTree struct {
	client curator.CuratorFramework
}

func NewZkTree(hosts []string, base string) (*ZkLiveTree, error) {
	client := curator.NewClient(hosts[0], curator.NewRetryNTimes(3, time.Second))

	if err := client.Start(); err != nil {
		return nil, err
	}

	if len(base) > 0 {
		if base[0] == '/' {
			base = base[1:]
		}

		client = client.UsingNamespace(base)
	}

	return &ZkLiveTree{client}, nil
}

// writes the in-memory ZK tree on to ZK server
func (t *ZkLiveTree) Write(tree ZkTree, force bool) error {
	return nil
}

// returns a list of actions after taking a diff of in-memory ZK tree and live ZK tree.
func (t *ZkLiveTree) Diff(tree ZkTree) (ZkActions, error) {
	return nil, nil
}

// performs create/delete/setvalue by executing a set of ZkActions on a live ZK tree.
func (t *ZkLiveTree) Execute(actions ZkActions, handler ZkActionHandler) error {
	return nil
}

func (t *ZkLiveTree) Node(znodePath string) (*ZkNode, error) {
	if data, err := t.client.GetData().ForPath(znodePath); err != nil {
		return nil, fmt.Errorf("fail to get data of node `%s`, %s", znodePath, err)
	} else if children, err := t.client.GetChildren().ForPath(znodePath); err != nil {
		return nil, fmt.Errorf("fail to get children of node `%s`, %s", znodePath, err)
	} else {
		var nodes []*ZkNode

		for _, child := range children {
			if node, err := t.Node(path.Join(znodePath, child)); err != nil {
				return nil, err
			} else {
				nodes = append(nodes, node)
			}
		}

		return &ZkNode{
			XMLName: xml.Name{
				Local: "zknode",
			},
			Name:     path.Base(znodePath),
			Value:    string(data),
			Children: nodes,
		}, nil
	}
}

func (t *ZkLiveTree) Root() (*ZkNode, error) {
	if children, err := t.client.GetChildren().ForPath("/"); err != nil {
		return nil, fmt.Errorf("fail to get children of root, %s", err)
	} else {
		var nodes []*ZkNode

		for _, child := range children {
			if node, err := t.Node(path.Join("/", child)); err != nil {
				return nil, err
			} else {
				nodes = append(nodes, node)
			}
		}

		return &ZkNode{
			XMLName: xml.Name{
				Local: "root",
			},
			Children: nodes,
		}, nil
	}
}

func (t *ZkLiveTree) Dump(depth int) (string, error) {
	return "", nil
}

func (t *ZkLiveTree) Xml() ([]byte, error) {
	if root, err := t.Root(); err != nil {
		return nil, err
	} else if data, err := xml.MarshalIndent(root, "", "  "); err != nil {
		return nil, err
	} else {
		return []byte(xml.Header + string(data)), nil
	}
}

type ZkLoadedTree struct {
	file *os.File
	root *ZkNode
}

func LoadZkTree(filename string) (*ZkLoadedTree, error) {
	if file, err := os.Open(filename); err != nil {
		return nil, fmt.Errorf("fail to open file `%s`, %s", filename, err)
	} else if data, err := ioutil.ReadFile(filename); err != nil {
		return nil, fmt.Errorf("fail to read file `%s`, %s", filename, err)
	} else {
		var node ZkNode

		if err := xml.Unmarshal(data, &node); err != nil {
			return nil, fmt.Errorf("fail to parse file `%s`, %s", filename, err)
		}

		return &ZkLoadedTree{
			file: file,
			root: &node,
		}, nil
	}
}

func (t *ZkLoadedTree) Execute(actions ZkActions, handler ZkActionHandler) error {
	return nil
}

func (t *ZkLoadedTree) Dump(depth int) (string, error) {
	return "", nil
}

func (t *ZkLoadedTree) String() (string, error) {
	return t.Dump(-1)
}

func (t *ZkLoadedTree) Xml() ([]byte, error) {
	if data, err := xml.MarshalIndent(t.root, "", "  "); err != nil {
		return nil, err
	} else {
		return []byte(xml.Header + string(data)), nil
	}
}

func (t *ZkLoadedTree) Diff(tree ZkTree) (ZkActions, error) {
	return nil, nil
}
