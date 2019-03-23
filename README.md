## 使用教程


---


#### 如何在程序中开始


```
XXX-service
    handlers
        middlewares.go
```

> #### middlewares.go

```golang
func WrapEndpoints(in svc.Endpoints) svc.Endpoints {
	result := prometheus.WrapEndpoint(in)
	out, ok := result.(svc.Endpoints)
	if ok {
		return out
	} else {
		logs.Info("add prometheus fail")
		return in
	}
}
```

> #### 然后再需要使用的地方进行Run


```golang
// start Prometheus
	go prometheus.Run(config.PrometheusAddr)
```


---

### 个性化



#### 如何定制自己的prometheus

- 重写prometheus代码
- 在原来的基础上再包上自己的中间件


#### 如何包装中间件


##### 通过init初始化需要监听的参数

```golang
func init() {
	fieldKeys := []string{"method", "error"}
	requestCount = kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "my_group",
		Subsystem: "string_service",
		Name:      "request_count",
		Help:      "Number of requests received.",
	}, fieldKeys)
	requestLatency = kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "my_group",
		Subsystem: "string_service",
		Name:      "request_latency_microseconds",
		Help:      "Total duration of requests in microseconds.",
	}, fieldKeys)
}

```



```
func WrapEndpoint(in interface{}) interface{} {
	// range members with reflect
	prometheusWrap := func(ep endpoint.Endpoint, epName string) endpoint.Endpoint {
		return func(ctx context.Context, req interface{}) (rep interface{}, err error) {
			defer func(begin time.Time) {
				lvs := []string{"method", epName, "error", fmt.Sprint(err != nil)}
			    //插入自己的中间件
                //requestCount.With(lvs...).Add(1)
				//requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
			}(time.Now())
			resp, err := ep(ctx, req)
			return resp, err
		}
	}
	vIn := reflect.ValueOf(in)
	newOne := reflect.New(reflect.TypeOf(in))
	vOut := reflect.Indirect(newOne)
	for i := 0; i < vIn.NumField(); i++ {
		endpoint := vIn.Field(i).Interface().(endpoint.Endpoint)
		endpoint = prometheusWrap(endpoint, vIn.Type().Field(i).Name)
		//logs.Info(vIn.Type().Field(i).Name + " is ok")
		vOut.Field(i).Set(reflect.ValueOf(endpoint))
	}
	return newOne.Elem().Interface()
}

```

#### 在原本的中间件上添加自己的内容
##### 只需要修改 middlewares.go


```golang

func WrapEndpoints(in svc.Endpoints) svc.Endpoints {
	result := prometheus.WrapEndpoint(in)
	out, ok := result.(svc.Endpoints)
	if ok {
		youOut,ok:=YouWrapEndpoint(out)
		if ok{
		    return youOut
		}else{
		    logs.Info("add your prometheus fail")
		    return in
		}
	} else {
		logs.Info("add prometheus fail")
		return in
	}
}

```


##### 或者直接使用自己的


```golang

func WrapEndpoints(in svc.Endpoints) svc.Endpoints {
	youOut,ok:=YouWrapEndpoint(in)
	if ok{
	    return youOut
	}else{
	    logs.Info("add your prometheus fail")
	    return in
	}
}

```


###### 编著人：Allen guo
###### 日期：  2018/9/28
