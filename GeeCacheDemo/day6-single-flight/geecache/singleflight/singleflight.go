package singleflight

import "sync"

// call 表示正在进行或者已经结束的请求
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

// Group map key to call
type Group struct {
	mu sync.Mutex // protect m
	m  map[string]*call
}

func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()         // 请求正在进行中，等待
		return c.val, c.err // 请求结束，返回
	}
	c := new(call)
	c.wg.Add(1)  // 发起请求前加锁
	g.m[key] = c // 添加到 g.m 表明对应的 key 已经有对应的请求在处理
	g.mu.Unlock()

	c.val, c.err = fn() // 调用fn, 发起请求
	c.wg.Done()         // 请求结束

	g.mu.Lock()
	delete(g.m, key) // 更新 g.m
	g.mu.Unlock()

	return c.val, c.err // 返回结果
}
