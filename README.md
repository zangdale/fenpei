# fenpei

简单的 http 转发服务

## install

```
go install github.com/zangdale/fenpei@latest
```

## conf

conf.json
```
{
    "debug": false,
    "port": 8899,
    "routers_file": "routers.json"
}
```
routers.json
```
{
    "port":7788,
    "routers":[
        {
            "router":"/a",
            "disable":false,
            "to":"http://127.0.0.1:8090"
        },
        {
            "router":"/b",
            "to":"http://127.0.0.1:8091"
        },
        {
            "router":"/c",
            "to":"http://127.0.0.1:8093"
        },
        {
            "forward_port":8094
        },
        {
            "router":"/",
            "dist_file":"./dist_d"
        }
    ]
}
```