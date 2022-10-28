package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type routersConf struct {
	Port    uint32   `json:"port"`
	Routers []router `json:"routers"`
}

type router struct {
	Router  *string `json:"router,omitempty"`
	Disable bool    `json:"disable"`

	To          *string `json:"to,omitempty"`
	ForwardPort *uint32 `json:"forward,omitempty"`
	DistFile    *string `json:"distFile,omitempty"`
}

func getRouters(fp string) *routersConf {
	c := &routersConf{
		Port:    7788,
		Routers: []router{},
	}
	b, err := os.ReadFile(fp)
	if err != nil {
		return c
	}
	err = json.Unmarshal(b, c)
	if err != nil {
		return c
	}

	// for i := 0; i < len(c.Routers); i++ {
	// 	if c.Routers[i].Disable {
	// 		c.Routers = append(c.Routers[:i], c.Routers[i+1:]...)
	// 	}
	// }
	return c

}

func confToRouters(routers []router) *BaseRouter {
	root := NewBaseRouter()
	for _, v := range routers {
		if v.Disable {
			continue
		}

		if v.Router != nil {
			if v.To != nil {
				root.AddRouter(*v.Router, HttpHandler(*v.Router, *v.To))
			} else {
				if v.DistFile != nil {
					root.AddRouter(*v.Router, FileHandler(*v.Router, *v.DistFile))
				}
			}
			continue
		}

		if v.ForwardPort != nil {
			root.AddRouter(*v.Router,
				HttpHandler(fmt.Sprintf("/%d", *v.ForwardPort),
					fmt.Sprintf("http://127.0.0.1:%d", *v.ForwardPort)))
		}

	}
	return root
}
